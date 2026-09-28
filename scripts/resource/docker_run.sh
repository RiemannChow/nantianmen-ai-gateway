#!/bin/sh

set -e
source ./init_config.sh

OLD_IFS="$IFS"
IFS=","
IFS="$OLD_IFS"


echo "" > config.yml

echo -e "mysql:" > config.yml
echo -e "  user_name: ${MYSQL_USER_NAME}" >> config.yml
echo -e "  password: ${MYSQL_PWD}" >> config.yml
echo -e "  ip: ${MYSQL_IP}" >> config.yml
echo -e "  port: ${MYSQL_PORT}" >> config.yml
echo -e "  db: ${MYSQL_DB}" >> config.yml
echo -e "redis:" >> config.yml
echo -e "  user_name: ${REDIS_USER_NAME}" >> config.yml
echo -e "  password: ${REDIS_PWD}" >> config.yml
echo -e "  addr: " >> config.yml
for s in $REDIS_ADDR
do
echo -e "    - $s" >> config.yml
done

echo -e "nsq:" >> config.yml
echo -e "  addr: ${NSQ_ADDR}" >> config.yml
echo -e "  topic_prefix: ${NSQ_TOPIC_PREFIX}" >> config.yml
echo -e "port: 8288" >> config.yml
echo -e "error_log:" >> config.yml
echo -e "  dir: ${ERROR_DIR}" >> config.yml
echo -e "  file_name: ${ERROR_FILE_NAME}" >> config.yml
echo -e "  log_level: ${ERROR_LOG_LEVEL}" >> config.yml
echo -e "  log_expire: ${ERROR_EXPIRE}" >> config.yml
echo -e "  log_period: ${ERROR_PERIOD}" >> config.yml

cat config.yml
nohup ./nantianmen-ai-gateway >> run.log 2>&1 &
wait_for_nantianmen_ai_gateway

nohup ./nantianmen-ai-gateway_ai_event_listen >> run.log 2>&1 &

if [[ ${Init} == "true" ]];then
  login_nantianmen_ai_gateway
  r=$(is_init)
  if [[ $r == "true" ]];then
    echo "Already initialized, skipping initialization."
  else
    wait_for_influxdb

      wait_for_apinto
      set_cluster

      wait_for_influxdb
      set_influxdb

      set_loki
      set_nsq
      set_openapi_config
      # 重启nantianmen-ai-gateway
      kill -9 $(pgrep nantianmen-ai-gateway)
      nohup ./nantianmen-ai-gateway >> run.log 2>&1 &
  fi
fi

tail -F run.log