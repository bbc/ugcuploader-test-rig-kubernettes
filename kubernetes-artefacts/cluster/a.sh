#!/usr/bin/env bash
H=$(aws iam get-policy --policy-arn arn:aws:iam::123456789012:policy/MySamplePolicy)
COMMAND=$(echo $H|tr -d '\n')
echo "|$COMMAND|"
