FROM docker.io/library/python:3.11-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1

# Install networking tools (net-tools for arp, iproute2 for ip commands)
RUN apt-get update && apt-get install -y \
    net-tools iproute2 && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy the application code
COPY main.py /app/

# Expose the application port
EXPOSE 12345

CMD ["python3", "main.py"]
