FROM python:3.9.10-alpine3.15

WORKDIR /opt/noisy

COPY requirements.txt .

RUN pip install -r requirements.txt

COPY . /opt/noisy

ENTRYPOINT ["python", "/opt/noisy/noisy.py"]

CMD ["--config", "config.json"]

