#!/bin/bash

# Controlla se è stata fornita una directory come argomento
if [ -z "$1" ]; then
  echo "Usage: $0 <directory>"
  exit 1
fi

# Legge il contenuto della directory
for file in "$1"/*; do
  # Controlla se è un file
  if [ -f "$file" ]; then
    #echo "./bin/alert.exe -f $file -d ./bin/tmp"
	./bin/alert.exe -f "$file" -d ./bin/tmp
  fi
done