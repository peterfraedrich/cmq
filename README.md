![CMQ logo](cmq.png)
> because even a bad queue is better than no queue at all

CMQ (config map queue) is a dumb idea of an application. For real, please don't use this in production. This takes a Kubernetes configMap and turns it into a queue. Why? Because I can and because its there. 

Inspired by ConsulMQ (github.com/peterfraedrich/consulmq)

Here's the latest benchmark for `Push()`-ing 100 items to the queue:
```
root@SATELLITE:~/dev/cmq# time ./cmq -external -debug
>>>>> DEBUG: &{Debug:true Kube:{IsExternal:true KubeconfigPath:/root/.kube/config Namespace:cmq ConfigMapName:default ShardSize:10 ShardPrefix:cmq} Server:{Host:0.0.0.0 Port:5000 UseSSL:false CertKeyPath:crt.key CertSecretPath:crt.secret} UI:{Enabled:true Port:8080}}
100

real    0m58.437s
user    0m0.442s
sys     0m0.086s
```

This why I say "PLEASE DON'T USE THIS". Its slow. Its Bad. I just wanted to see if I could.