package main

import (
    "io"
    "log"
    "net"
    "net/http"
    "os"
    "time"
    "sync"
    "strings"
    "crypto/tls"
    "github.com/miekg/dns"
    "encoding/base64"
)

var debugOn bool
var hostsDNS sync.Map
var tcg *tls.Config

func main() {
    argc:=len(os.Args)
    if argc <2 {
        log.Println("superhosts: server 127.0.0.1:8081 hosts debug=on/off(Optional,default=off)")
        return
    }
    
    if argc ==4 {
        if os.Args[3]=="debug=on"{
            debugOn=true
            log.Println("Debug: debug on")
        }
    }
    ipaddr:=os.Args[1]

    cert, err := tls.LoadX509KeyPair("doh.crt", "doh.key")
    if err != nil {
        log.Println("need doh.crt/doh.key:",err)
        return
    }
    tcg = &tls.Config{
        Certificates: []tls.Certificate{cert},
    }
    
    server := &http.Server{
        Addr: ipaddr,
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.Method == http.MethodConnect {
                handleConnect(w, r)
            } else {
                handleHTTP(w, r)
            }
        }),
    }
    
    hosts_data, err := os.ReadFile(os.Args[2])
    if err != nil {
        log.Fatalf("ERR: Read hosts fail: %v", err)
    }
    les := strings.Split(string(hosts_data), "\n")
    for _, le := range les {
        if le != "" && le[0]!='#' && le[0]!=';' {
        rt := strings.ReplaceAll(le, "\r", "")
        pt := strings.SplitN(rt, " ", 2)
        if len(pt) == 2 {
            pr := strings.TrimSpace(pt[1])
            sr := strings.TrimSpace(pt[0])
            _,ok:=hostsDNS.Load(pr)
            if !ok {
            hostsDNS.Store(pr,sr)
            } else {
            hostsDNS.Store(pr+"1",sr)
            }
            log.Println("INFO: load super hosts", rt)
        } else {
            log.Println("ERR: load format err", rt)
        }
        }
    }
    
    log.Println("INFO: server listening on:",ipaddr)
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}

func nslookup(name string, port uint16,dnsServer string) ([]byte, error) {
    msg := dns.Msg{}
    msg.SetQuestion(dns.Fqdn(name), port)
    msg.RecursionDesired = true
    dnsPack,err:=msg.Pack()
    if err != nil { return nil, err }
    conn, err := net.Dial("udp", dnsServer)
    if err != nil { return nil, err }
    defer conn.Close()
    conn.SetDeadline(time.Now().Add(5*time.Second))
    if _, err := conn.Write(dnsPack); err != nil { return nil, err }
    buf := make([]byte, 4096)
    n, err := conn.Read(buf)
    if err != nil { return nil, err }
    return buf[:n], nil
}
func handleConnect(w http.ResponseWriter, r *http.Request) {
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "connect not supported", http.StatusInternalServerError)
        return
    }

    oc, _, err := hijacker.Hijack()
    if err != nil {
        if debugOn {
        log.Println("Debug: error:", err)
        }
        return
    }
    
    domain, port, err := net.SplitHostPort(r.Host)
    if err != nil {
        domain = r.Host
    }
    hosts :=""
    dnsval,ok:=hostsDNS.Load(domain)
    if ok {
        hosts = dnsval.(string)
    }
    msg := new(dns.Msg)
    var or net.Conn
    doH :=false
    if hosts !="" {
        if debugOn {
        log.Println("Debug: find IP:",r.Host,hosts)
        }
        rip:=""
        if strings.Contains(hosts,"@") || strings.Contains(hosts,"!") {
            if hosts[0]=='@'{
            if len(hosts)>2 && hosts[1]=='!'{
            cname:=strings.TrimPrefix(hosts, "@!")
            dnsval,ok =hostsDNS.Load(cname)
                if ok {
            dnsServer:=dnsval.(string)
            if ok {
            if len(dnsServer)>2 && dnsServer[0]=='!'{
            dnsServer=strings.TrimPrefix(dnsServer, "!dns=")
            answers, err := nslookup(cname,65,dnsServer)
            if err != nil {
                log.Println("ERR: dns lookup fail: %v", err)
                oc.Close()
            return}
            if msg.Unpack(answers) != nil {
            log.Println("ERR: DNS response: %v\n", err)
                oc.Close()
        return
        }
        doH=true
        dnsval,ok =hostsDNS.Load(domain+"1")
                    if ok {
        rip=dnsval.(string)
        } else {
        answer,err:=nslookup(cname,dns.TypeA,dnsServer)
        if err != nil {
            log.Println("ERR: dns lookup fail: %v", err)
            oc.Close()
        return
        }
        rev := new(dns.Msg)
            if rev.Unpack(answer) != nil {
            log.Println("ERR: DNS response: %v\n", err)
            oc.Close()
        return
        }
        for _, rr := range rev.Answer {
        switch v := rr.(type) {
            case *dns.A:
                    rip = v.A.String()
            case *dns.AAAA:
                rip = v.AAAA.String()
            }}}}}}
            } else {
            cname:=strings.TrimPrefix(hosts, "@")
            dnsval,ok =hostsDNS.Load(cname)
                    if ok {
            rip = dnsval.(string)
            }}
            } else {
            dnsServer:=strings.TrimPrefix(hosts, "!dns=")
            answers,err:=nslookup(domain,dns.TypeA,dnsServer)
            if err != nil {
                    log.Println("ERR: dns lookup fail: %v", err)
            oc.Close()
            return
        }
                if msg.Unpack(answers) != nil {
            log.Println("ERR: DNS response: %v\n", err)
            oc.Close()
                    return
        }
        for _, rr := range msg.Answer {
                switch v := rr.(type) {
            case *dns.A:
        rip = v.A.String()
        case *dns.AAAA:
                rip = v.AAAA.String()
        }}}
        } else {
            rip = hosts
        }
        or, err = net.Dial("tcp", rip+":"+port)
    } else {
    or, err = net.Dial("tcp", r.Host)
    }
    
    if err != nil {
        if debugOn {
       log.Println("Debug: Dial:", err)
    }
     oc.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
        oc.Close()
    return
        }
    oc.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
    var l net.Conn;var I net.Conn
    if doH {
    tlsConn := tls.Server(oc, tcg)
    err = tlsConn.Handshake()
        if err != nil {
    oc.Close()
    or.Close()
    log.Println("TLS Handshake failed:", err)
            return
    }
    dohrecv := ""
    for _, ans := range msg.Answer {
        if httpsRecord, ok := ans.(*dns.HTTPS); ok {
    for _, i := range httpsRecord.Value {
                switch v := i.(type) {
        case *dns.SVCBECHConfig:
                    dohrecv = base64.StdEncoding.EncodeToString(v.ECH)
        break
            case *dns.SVCBIPv4Hint:
                    v.Hint[0].String()
        }}
        } else {
            log.Println("ERR: DNS recv err");}
    }
    
    if len(dohrecv) == 0 {
        log.Println("ERR: DNS recv err, change dns server")
            oc.Close()
            or.Close()
            return
    }
        dohsrv, err := base64.StdEncoding.DecodeString(dohrecv)
        if err != nil {
        oc.Close()
        or.Close()
        return
    }
    c := &tls.Config{
        MinVersion:                     tls.VersionTLS13,
    ServerName:                     domain,
        EncryptedClientHelloConfigList: dohsrv,
    }
    tlsconn := tls.Client(or, c)
            err = tlsconn.Handshake()
    if err != nil {
        or.Close()
            tlsConn.Close()
        log.Println("ERR: connect tls", err)
    return
    }
    l=tlsConn;I=tlsconn;} else {l=oc;I=or;}
        ends := make(chan bool, 2)
    go func() {
        io.Copy(I, l)
        ends <- true
    }()

    go func() {
        io.Copy(l, I)
        ends <- true
    }()
    
    <-ends
    l.Close()
    I.Close()
    <-ends
}
func handleHTTP(w http.ResponseWriter, r *http.Request) {
    if r.URL.Scheme == "" {
        r.URL.Scheme = "http"
    }
    if r.URL.Host == "" {
        r.URL.Host = r.Host
    }

    resp, err := http.DefaultTransport.RoundTrip(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    defer resp.Body.Close()
    for k, vv := range resp.Header {
        for _, v := range vv {
            w.Header().Add(k, v)
        }
    }
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}