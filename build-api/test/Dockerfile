FROM isldocker/build:v5

RUN eval `ssh-agent -s` && \
    ssh-add ~/.ssh/id_rsa && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts && \
    git clone git@github.com:istrategylabs/franklin-test.git

RUN cd franklin-test && git pull origin 
RUN cd franklin-test && git checkout 7b3af09db495427d067704b5e9290d8b81d682c5

RUN cd franklin-test && npm install
RUN cd franklin-test && npm run build
