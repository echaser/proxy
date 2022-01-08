package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type Proxy struct{}

// 当浏览器需要通过代理服务器发起HTTPS请求时，由于请求的站点地址和端口号都是加密保存于HTTPS请求头中的，
// 代理服务器自身无法读取通信内容，怎么知道该往哪里发送请求呢？
// 为了解决这个问题，浏览器先通过明文HTTP形式向代理服务器发送一个CONNECT请求告诉它目标地址及端口号。
// 当代理服务器收到这个请求后，会使用地址和端口与目标站点建立一个TCP连接，
// 连接建立成功后返回一个HTTP 200状态码告诉浏览器与该站点的加密通道已建成。
// 接下来代理服务器仅仅是来回传输浏览器与该服务器之间的加密数据包，代理服务器并不需要解析这些内容以保证HTTPS的安全性
func (p *Proxy) Serve(w http.ResponseWriter, r *http.Request) {
	// 只有当浏览器配置为使用代理服务器时才会用到CONNECT方法
	if r.Method == http.MethodConnect {
		p.httpsHandler(w, r)
	} else {
		p.httpHandler(w, r)
	}
}

func (*Proxy) httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("received HTTP request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
	transport := http.DefaultTransport
	// X-Forwarded-For https://www.runoob.com/w3cnote/http-x-forwarded-for.html
	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior, ok := r.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		r.Header.Set("X-Forwarded-For", clientIP)
	}

	// RoundTrip 不会修改Request
	resp, err := transport.RoundTrip(r)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (*Proxy) httpsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("received HTTPS request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
	// 与目标服务器建立连接
	serverConn, err := net.DialTimeout("tcp", r.Host, time.Second*30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	// 告诉客户端连接成功
	w.WriteHeader(http.StatusOK)

	// 接管连接
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	go copyData(clientConn, serverConn)
	go copyData(serverConn, clientConn)
}
