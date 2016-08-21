package main

import (
	"github.com/456vv/vforward/v1"
    "flag"
    "net"
    "time"
    "log"
    "fmt"
)

var fNetwork = flag.String("Network", "tcp", "网络地址类型")

var fListen = flag.String("Listen", "", "本地网卡监听地址 (format \"0.0.0.0:123\")")

var fFromLocal = flag.String("FromLocal", "0.0.0.0", "转发请求的源地址")
var fToLemote = flag.String("ToLemote", "", "转发请求的目地址 (format \"22.23.24.25:234\")")

var fTimeout = flag.Duration("Timeout", time.Second*5, "转发连接时候，请求远程连接超时。单位：ns, us, ms, s, m, h")
var fMaxConn = flag.Int("MaxConn", 500, "限制连接最大的数量")
var fReadBufSize = flag.Int("ReadBufSize", 4096, "交换数据缓冲大小。单位：字节")

//commandline:l2d-main.exe -Listen 127.0.0.1:1201 -ToLemote 127.0.0.1:1202
func main(){
    flag.Parse()
    if flag.NFlag() == 0 {
        flag.PrintDefaults()
        return
    }
    var err error
    if *fListen == "" || *fToLemote == "" {
        log.Printf("地址未填，本地监听地址 %q, 转发到远程地址 %q", *fListen, *fToLemote)
        return
    }

    listenIP, err := net.ResolveTCPAddr(*fNetwork, *fListen)
    if err != nil {
        log.Println(err)
        return
    }
    dialIP, err := net.ResolveTCPAddr(*fNetwork, *fToLemote)
    if err != nil {
        log.Println(err)
        return
    }

    dial := &vforward.Addr{
        Network:*fNetwork,
        Local: &net.TCPAddr{IP: net.ParseIP(*fFromLocal),Port: 0,},
        Remote: dialIP,
    }
    listen := &vforward.Addr{
        Network:*fNetwork,
        Local: listenIP,
    }

    ld := &vforward.L2D{
        MaxConn: *fMaxConn,            // 限制连接最大的数量
        ReadBufSize: *fReadBufSize,    // 交换数据缓冲大小
        Timeout: *fTimeout,            // 发起连接超时
    }

    lds, err := ld.Transport(dial, listen)
    if err != nil {
        log.Println(err)
        return
    }
    exit := make(chan bool, 1)
    go func(){
        defer func(){
            lds.Close()
            exit <- true
            close(exit)
        }()
        log.Println("L2D启动了")

        var in0 string
        for err == nil  {
            log.Println("输入任何字符可以退出L2D!")
            fmt.Scan(&in0)
            if in0 != "" {
                return
            }
        }
    }()
    defer lds.Close()
    err = lds.Swap()
    if err != nil {
        log.Println("错误：%s", err)
    }
    <-exit
    log.Println("L2D退出了")
}