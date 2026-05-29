FROM ubuntu:24.04

RUN apt-get update && apt-get install -y \
    curl git unzip tar ca-certificates xz-utils \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -s /bin/bash tester
USER tester
ENV PATH="/home/tester/.local/bin:${PATH}"
RUN mkdir -p /home/tester/.local/bin

COPY --chown=tester:tester devops-starter /home/tester/.local/bin/
COPY --chown=tester:tester verify /home/tester/.local/bin/

ENTRYPOINT ["verify"]
