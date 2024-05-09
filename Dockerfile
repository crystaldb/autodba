# Dockerfile to containerize AutoDBA agent

# Use an official Python runtime as a parent image
FROM python:3.12.3-slim as base

# Set environment varibles
ENV PYTHONDONTWRITEBYTECODE 1
ENV PYTHONUNBUFFERED 1

# Set work directory
WORKDIR /app

# Install Python dependencies
RUN pip install --upgrade pip
COPY requirements.txt /app/
RUN pip install -r requirements.txt

# Copy the current directory contents into the container
COPY ./src /app/

# Expose port 8080 for HTTP traffic
EXPOSE 8080

# Start the AutoDBA Agent when the container launches
CMD ["python", "-m", "AutoDBA.agent"]
