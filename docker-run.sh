docker run --rm \
    --name fhaas \
    -v $(pwd)/src/logs:/logs \
    -v /home:/home \
    -p 7500:8080 \
    alantelles/fhaas --authurl=http://192.168.43.30:8000/api/v1/auth
