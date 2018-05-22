git pull

chmod a+x start-agent.sh

docker stop provider-small
docker rm provider-small
docker stop provider-medium
docker rm provider-medium
docker stop provider-large
docker rm provider-large
docker stop consumer
docker rm consumer
#docker run   -p 30000:30000 --name provider-small -v /www/middlewar/mydemo/mesh-agnet/mesh-agent/target/mesh-agent-1.0-SNAPSHOT.jar:/root/dists/mesh-agent.jar  middlewar/agent  provider-amal
#  --cpu-period 50000 --cpu-quota 60000 -m 4g
docker run -d --cpu-period 50000 --cpu-quota 90000 -m 6g  --network host  --name provider-large -v /www/middlewar/mydemo/go-crazy/docker/start-agent.sh:/usr/local/bin/start-agent.sh   -v /www/middlewar/mydemo/go-crazy/docker/server.exe:/root/workspace/agent/server.exe middlewar/agent:1.0  provider-large
docker run -d --cpu-period 50000 --cpu-quota 60000 -m 4g  --network host --name provider-medium -v /www/middlewar/mydemo/go-crazy/docker/start-agent.sh:/usr/local/bin/start-agent.sh   -v /www/middlewar/mydemo/go-crazy/docker/server.exe:/root/workspace/agent/server.exe middlewar/agent:1.0  provider-medium
docker run -d --cpu-period 50000 --cpu-quota 30000 -m 2g  --network host --name provider-small  -v /www/middlewar/mydemo/go-crazy/docker/start-agent.sh:/usr/local/bin/start-agent.sh   -v /www/middlewar/mydemo/go-crazy/docker/server.exe:/root/workspace/agent/server.exe middlewar/agent:1.0  provider-small
docker run --cpu-period 50000 --cpu-quota 200000 -m 4g  --network host --name consumer          -v /www/middlewar/mydemo/go-crazy/docker/start-agent.sh:/usr/local/bin/start-agent.sh   -v /www/middlewar/mydemo/go-crazy/docker/server.exe:/root/workspace/agent/server.exe middlewar/agent:1.0  consumer