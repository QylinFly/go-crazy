# Builder container
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/services AS builder


# Runner container
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/debian-jdk8

# 基础拷贝
COPY --from=builder /root/workspace/services/mesh-provider/target/mesh-provider-1.0-SNAPSHOT.jar /root/dists/mesh-provider.jar
COPY --from=builder /root/workspace/services/mesh-consumer/target/mesh-consumer-1.0-SNAPSHOT.jar /root/dists/mesh-consumer.jar
COPY --from=builder /usr/local/bin/docker-entrypoint.sh /usr/local/bin


# 拷贝程序和配置
COPY . /root/workspace/agent
# 配置文件
COPY ./.env.yml  /root/workspace/agent
COPY start-agent.sh /usr/local/bin

RUN set -ex \
    && chmod a+x /usr/local/bin/start-agent.sh \
    # golang 程序
    && chmod a+x /root/workspace/agent/server.exe \  
    && mkdir -p /root/logs


EXPOSE 8087
ENTRYPOINT ["docker-entrypoint.sh"]
