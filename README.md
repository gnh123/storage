# storage
存储引擎，第一个版本依据海量小文件存储, 这是一个库，可以直接嵌入到程序中使用。目前版本不建议在生产中使用，也自带了一个很简易的服务端，方便压测性能

# 运行服务端
```
./storage server -d ./my-store -s 64GB
````
# 运行压测
```
./storage benchmark -p -b "hello world" -d 20s -s 127.0.0.1:8080
  Completed           99870 requests [2022-10-30 23:45:53.946218]
  Completed          197209 requests [2022-10-30 23:45:55.946235]
  Completed          297909 requests [2022-10-30 23:45:57.946233]
  Completed          399386 requests [2022-10-30 23:45:59.946219]
  Completed          501585 requests [2022-10-30 23:46:01.946443]
  Completed          588044 requests [2022-10-30 23:46:03.946283]
  Completed          687964 requests [2022-10-30 23:46:05.946247]
  Completed          786152 requests [2022-10-30 23:46:07.946261]
  Completed          883646 requests [2022-10-30 23:46:09.946316]
  Completed          979513 requests [2022-10-30 23:46:11.946272]
  Finished           979523 requests


Status Codes:            979523:200  [count:code]
Concurrency Level:      10
Time taken for tests:   20.000225667s
Complete requests:      979523
Failed requests:        0
Total Read Data:        170860963 bytes
Total Read body         50379634 bytes
Total Write Body        10774753 bytes
Requests per second:    48975.59739119318 [#/sec] (mean)
Time per request:       0.02041833184825675 [ms] (mean)
Time per request:       0.20418331848256754 [ms] (mean, across all concurrent requests)
Transfer rate:          8342.726325083295 [Kbytes/sec] received
Percentage of the requests served within a certain time (ms)
  50%    173.375µs
  66%    206.5µs
  75%    233.833µs
  80%    258.541µs
  90%    341.625µs
  95%    406.792µs
  98%    533.459µs
  99%    664.417µs
 100%    61.231584ms
```
