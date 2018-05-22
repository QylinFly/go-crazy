#!/bin/bash

# Old Version
# ETCD_HOST=$(ip addr show docker0 | grep 'inet\b' | awk '{print $2}' | cut -d '/' -f 1)
# ETCD_PORT=2379

# if [ "$ETCD_HOST"x == ""x ]
# then
#    ETCD_URL=http://172.17.0.1:$ETCD_PORT
# else
#    ETCD_URL=http://$ETCD_HOST:$ETCD_PORT
# fi

# docker 
ETCD_HOST=etcd
ETCD_PORT=2379
# ETCD_URL=http://$ETCD_HOST:$ETCD_PORT

ETCD_URL=http://172.17.0.1:$ETCD_PORT

echo ETCD_URL = $ETCD_URL

if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."
  cd /root/workspace/agent/
    ./server.exe -Dtype=consumer -Dserver.port=20000  -Detcd.url=$ETCD_URL  -Dlogs.dir=/root/logs 

elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."

    cd /root/workspace/agent/
    ./server.exe -Dtype=provider -Ddubbo.protocol.port=20880 -DChannels=50 -Dserver.port=30000  -Detcd.url=$ETCD_URL  -Dlogs.dir=/root/logs 

  # java -jar \
  #      -Xms512M \
  #      -Xmx512M \
  #      -Dmy.io.threads=2\
  #      -Dmy.io.workers=80\
  #      -Dmy.nio.threads=2\
  #      -Dmy.nio.channels=70\
  #      -Dtype=provider \
  #      -Dserver.port=30000\
  #      -Ddubbo.protocol.port=20889 \
  #      -Detcd.url=$ETCD_URL \
  #      -Dlogs.dir=/root/logs \
  #      /root/dists/mesh-agent.jar
elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."

    cd /root/workspace/agent/
    ./server.exe -Dtype=provider -Ddubbo.protocol.port=20880 -DChannels=100 -Dserver.port=30001  -Detcd.url=$ETCD_URL  -Dlogs.dir=/root/logs 

  # java -jar \
  #      -Xms1536M \
  #      -Xmx1536M \
  #      -Dmy.io.threads=3\
  #      -Dmy.io.workers=100\
  #      -Dmy.nio.threads=3\
  #      -Dmy.nio.channels=90\
  #      -Dtype=provider \
  #      -Dserver.port=30001\
  #      -Ddubbo.protocol.port=20890 \
  #      -Detcd.url=$ETCD_URL \
  #      -Dlogs.dir=/root/logs \
  #      /root/dists/mesh-agent.jar
elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."
    cd /root/workspace/agent/
    ./server.exe -Dtype=provider -Ddubbo.protocol.port=20880 -DChannels=180 -Dserver.port=30002  -Detcd.url=$ETCD_URL  -Dlogs.dir=/root/logs 


  # java -jar \
  #      -Xms2560M \
  #      -Xmx2560M \
  #      -Dmy.io.threads=5\
  #      -Dmy.io.workers=210\
  #      -Dmy.nio.threads=4\
  #      -Dmy.nio.channels=201\
  #      -Dtype=provider \
  #      -Dserver.port=30002\
  #      -Ddubbo.protocol.port=20891 \
  #      -Detcd.url=$ETCD_URL \
  #      -Dlogs.dir=/root/logs \
  #      /root/dists/mesh-agent.jar
else
  echo "Unrecognized arguments, exit."
  exit 1
fi
