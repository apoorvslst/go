#!/bin/bash
# ============================================================
# Lecture 2: Docker + PostgreSQL Commands
# Ye saare commands video mein use hue the
# ============================================================

# -------- IMAGE DOWNLOAD --------
# PostgreSQL ka image download karo (alpine = chhota size)
docker pull postgres:12-alpine

# Check karo image aa gayi ya nahi
docker images

# -------- CONTAINER START --------
# PostgreSQL container start karo
docker run --name postgres12 \
  -p 5432:5432 \
  -e POSTGRES_USER=root \
  -e POSTGRES_PASSWORD=secret \
  -d postgres:12-alpine

# -------- CHECK STATUS --------
# Running containers dekho
docker ps

# Saare containers dekho (band bhi)
docker ps -a

# Container ke logs dekho
docker logs postgres12

# -------- CONTAINER KE ANDAR --------
# PostgreSQL console mein jaao
docker exec -it postgres12 psql -U root

# Container ka shell kholo
docker exec -it postgres12 /bin/sh

# Database banao (container ke andar se)
# docker exec -it postgres12 /bin/sh ke baad:
#   createdb --username=root --owner=root simple_bank
#   psql simple_bank
#   SELECT now();
#   \q
#   exit

# Database banao (bahar se directly)
docker exec -it postgres12 createdb --username=root --owner=root simple_bank

# Database delete karo
docker exec -it postgres12 dropdb simple_bank

# -------- CONTAINER MANAGE --------
# Container band karo
docker stop postgres12

# Container phir se chalu karo
docker start postgres12

# Container permanently delete karo (pehle stop karo)
docker stop postgres12
docker rm postgres12
