FROM python:3.11-slim

ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY main.py /app/

RUN pip install requests

CMD ["python3", "main.py"]
