FROM nginx:latest

# Copy the template file into the container
COPY index.html.template /usr/share/nginx/html/index.html.template

# Environment variable 
ENV API_SERVER_URL="http://localhost:8080/employees"

# Inster the placeholders $API_SERVER_URL with environment variable values
CMD /bin/bash -c "envsubst < /usr/share/nginx/html/index.html.template > /usr/share/nginx/html/index.html && nginx -g 'daemon off;'"