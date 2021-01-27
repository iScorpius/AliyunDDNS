# AliyunDDNS
Aliyun DDNS. 阿里云动态域名解析

## 编译
在go语言环境下编译

## 查看帮助
```
> aliyunddns -h
  -id
        阿里云AccessKeyId

  -secret
        阿里云AccessKeySecret

  -domain
        域名名称(example.com)

  -rr
        主机记录值(www)

  -type
        (默认值: A) 解析记录类型(A, NS, MX, TXT, CNAME, SRV, AAAA, CAA, etc...)

  -priority
        (默认值: 1) MX记录优先级

  -ttl
        (默认值: 600) 解析生效时间

  -dns
        (默认值: 223.5.5.5) DNS,用于判断解析是否生效
```

## 使用示例
```
> aliyunddns -id="your access key id" -secret="your access key secret" -domain="your domian" -rr="record" type="record type"
```
