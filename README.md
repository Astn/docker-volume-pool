# docker volume-pool plugin
goal: add support for something similar to docker -P but for volumes. So docker -V would consistantly automount volumes for a given image and container name.

Slacker @ https://dockerdenver.slack.com/

ex:
```
docker run -v 0:/data0 -v 1:/data1  --plugin=volume-pool --volumeroot=/foo [--volume-allocation-scheme=(ordered,round-robin,least-allocated,least-iops)]  [--volume-no-cache] --name elasticsearch elasticsearch
```
given a dir at
with folders mounted to devices
```
/foo
   /jboddisk0
   /jboddisk1
   /jboddisk2
   /jboddisk3
   /jboddisk4
   /jboddisk5
```
will create on the file system:
```
/foo
   /jboddisk0
      /elasticsearch-elasticsearch-data0
   /jboddisk1
      /elasticsearch-elasticsearch-data1
   /jboddisk2
   /jboddisk3
   /jboddisk4
   /jboddisk5
```
will create a docker volume container with volumes for key(imagename:containername)
```
  -v /foo/jboddisk0/elasticsearch-elasticsearch-data0:data0
  -v /foo/jboddisk1/elasticsearch-elasticsearch-data1:data1  
```
and bind that volume container to the elasticsearch container.  
