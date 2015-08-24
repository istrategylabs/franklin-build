FROM python:2.7
ENV PYTHONUNBUFFERED 1
RUN mkdir /code
WORKDIR /code
ADD build-api/requirements.txt /code/
RUN pip install -r requirements.txt
ADD . /code/
