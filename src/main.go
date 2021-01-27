package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

type Auth struct {
	AccessKeyId     string
	AccessKeySecret string
}

type Config struct {
	Domain   string
	RR       string
	Type     string
	Value    string
	Priority int64
	TTL      int64
	Host     string
}

var auth = Auth{}
var config = Config{}

func main() {
	argument()
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", auth.AccessKeyId, auth.AccessKeySecret)
	ip := queryRealIP()
	config.Value = ip
	domainIPCheck(ip)
	record, err := describeSubDomainRecords(client)
	if err != nil {
		addDomainRecord(client)
		return
	}
	updateDomainRecord(client, record)
}

func argument() {
	flag.StringVar(&auth.AccessKeyId, "id", "", "阿里云AccessKeyId")
	flag.StringVar(&auth.AccessKeySecret, "secret", "", "阿里云AccessKeySecret")
	flag.StringVar(&config.Domain, "domain", "", "域名名称(example.com)")
	flag.StringVar(&config.RR, "rr", "", "主机记录值(www)")
	flag.StringVar(&config.Type, "type", "A", "解析记录类型(A, NS, MX, TXT, CNAME, SRV, AAAA, CAA, etc...)")
	flag.Int64Var(&config.Priority, "priority", 1, "MX记录优先级")
	flag.Int64Var(&config.TTL, "ttl", 600, "解析生效时间")
	flag.StringVar(&config.Host, "dns", "223.5.5.5", "DNS,用于判断解析是否生效")
	flag.Usage = func() {
		order := []string{"id", "secret", "domain", "rr", "type", "priority", "ttl", "dns"}
		flagSet := flag.CommandLine
		for _, name := range order {
			flag := flagSet.Lookup(name)
			fmt.Printf("  -%s\r\n", flag.Name)
			if flag.DefValue != "" {
				fmt.Printf("        (默认值: %s) %s\r\n\r\n", flag.Value, flag.Usage)
			} else {
				fmt.Printf("        %s\r\n\r\n", flag.Usage)
			}
		}
	}
	flag.Parse()
}

func domainIPCheck(ip string) {
	var out bytes.Buffer
	nslookup := exec.Command("nslookup", config.RR+"."+config.Domain, config.Host)
	nslookup.Stdout = &out
	err := nslookup.Run()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	reg := regexp.MustCompile(`\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	for _, value := range reg.FindAll(out.Bytes(), -1) {
		if ip == string(value) {
			os.Exit(0)
		}
	}
}

func describeSubDomainRecords(client *alidns.Client) (alidns.Record, error) {
	request := alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"

	request.SubDomain = config.RR + "." + config.Domain
	response, err := client.DescribeSubDomainRecords(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	if len(response.DomainRecords.Record) != 0 {
		return response.DomainRecords.Record[0], nil
	}
	return alidns.Record{}, errors.New("record is nil")
}

func addDomainRecord(client *alidns.Client) (string, string) {
	request := alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"

	request.DomainName = config.Domain
	request.RR = config.RR
	request.Type = config.Type
	request.Value = config.Value
	request.Priority = requests.Integer(strconv.FormatInt(config.Priority, 10))
	request.TTL = requests.Integer(strconv.FormatInt(config.TTL, 10))

	response, err := client.AddDomainRecord(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	return response.RecordId, response.RequestId
}

func updateDomainRecord(client *alidns.Client, record alidns.Record) (string, string) {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = record.RecordId
	request.RR = config.RR
	request.Type = config.Type
	request.Value = config.Value
	request.Priority = requests.Integer(strconv.FormatInt(config.Priority, 10))
	request.TTL = requests.Integer(strconv.FormatInt(config.TTL, 10))

	response, err := client.UpdateDomainRecord(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	return response.RecordId, response.RequestId
}

func queryRealIP() string {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://whatismyip.akamai.com")
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	return string(body)
}
