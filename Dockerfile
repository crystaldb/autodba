# Use the official PostgreSQL image as the base image
FROM postgres:16.3

# Set environment varibles
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

# Install netcat, Supervisor, and Python
RUN apt-get update && \
    apt-get install -y netcat-traditional supervisor python3 python3-pip python3-venv && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Create a Python virtual environment
RUN python3 -m venv /opt/venv

# Activate virtual environment
ENV PATH="/opt/venv/bin:$PATH"

# create the appropriate directories
ENV HOME=/home/app
ENV APP_HOME=/home/app
WORKDIR $APP_HOME

# Create a directory for the application
RUN mkdir -p $APP_HOME

# create the app user and group
RUN addgroup --system pgautodba_user && adduser --system --group pgautodba_user

# Set the working directory in the container
WORKDIR $APP_HOME

# Install Python dependencies
RUN pip install --upgrade pip
COPY requirements.txt $APP_HOME/
RUN pip3 install --no-cache-dir -r requirements.txt

# Copy the current directory contents into the container
COPY ./src $APP_HOME/
RUN mkdir -p $APP_HOME/logs

# Copy the Supervisor configuration file
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose port 8080 for HTTP traffic
EXPOSE 8080

# chown all the files to the app user
RUN chown -R pgautodba_user:pgautodba_user $APP_HOME

# run entrypoint.sh
CMD ["/home/app/entrypoint.sh"]
