# vim: ft=dockerfile
FROM public.ecr.aws/ubuntu/ubuntu:20.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install --no-install-recommends --no-install-suggests -y curl ca-certificates

# 本番では xbuild で 3.0.2 を入れるのでだいじょうぶです
RUN mkdir -p /usr/local/share/keyrings
RUN curl -LSsfo /usr/local/share/keyrings/sorah-ruby.asc https://sorah.jp/packaging/debian/3F0F56A8.pub.txt
RUN echo "deb [signed-by=/usr/local/share/keyrings/sorah-ruby.asc] http://cache.ruby-lang.org/lab/sorah/deb/ focal main" > /etc/apt/sources.list.d/sorah-ruby.list \
  && apt-get update \
  && apt-get upgrade -y \
  && apt-get install --no-install-recommends --no-install-suggests -y \
  ruby \
  ruby-dev \
  ruby3.0 \
  ruby3.0-dev \
  libruby3.0 \
  ruby3.0-gems \
  default-mysql-client \
  libmariadbclient-dev \
  build-essential \
  zlib1g-dev \
  tzdata

RUN ln -sf /usr/share/zoneinfo/Asia/Tokyo /etc/localtime

COPY webapp/public /webapp/public

WORKDIR /webapp/ruby
COPY webapp/ruby/Gemfile* /webapp/ruby/
RUN bundle install --jobs 300

ENV LANG=C.UTF-8
