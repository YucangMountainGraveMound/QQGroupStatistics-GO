FROM docker.elastic.co/elasticsearch/elasticsearch:6.3.1
MAINTAINER Dormon Zhou "dormon@dormon.net"

# IK 分词器
RUN /usr/share/elasticsearch/bin/elasticsearch-plugin install --batch https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v6.3.1/elasticsearch-analysis-ik-6.3.1.zip

# 拼音分词器
RUN /usr/share/elasticsearch/bin/elasticsearch-plugin install --batch https://github.com/medcl/elasticsearch-analysis-pinyin/releases/download/v6.3.1/elasticsearch-analysis-pinyin-6.3.1.zip