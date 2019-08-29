# GO gRPC 


ls, cp, and find functionality from server to client via gRPC streaming with protobuf defined api

./client -op=ls -file=/ returns server listing of root  i.e. ls -la /  
./client -op=ls -file=/usr returns server /usr
./client -op=ls -file=/home/USER  etc...

./client -op=cp -file=/etc/systemd/system.conf -dest=system.conf
     copy servers system.conf to current folder
     
./client -op=find -file=fileToFind 
      return full path to fileToFind on server
      
./client -op=find -file=/fileToFind    start search at root
      return full path to fileToFind on server      
      
./client -op=find -file=/home/fileToFind    start search at /home
      return full path to fileToFind on server            






