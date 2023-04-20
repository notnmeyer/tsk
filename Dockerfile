FROM scratch
COPY dist/tsk_linux_amd64_v1/tsk .
ENTRYPOINT ["/tsk"]