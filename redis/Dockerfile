# 使用官方的 Redis 镜像作为基础镜像
FROM redis:7.4.0

# 设置 Redis 的配置文件路径
COPY ./config/redis.conf /usr/local/etc/redis/redis.conf

# 通过配置文件启动 Redis
CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]

# 暴露 Redis 的默认端口
EXPOSE 6379
