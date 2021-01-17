
counter=1000
while [ $counter -le 1100 ]
do
printf "\nLoad step:${counter}\n"
uri="http://127.0.0.1:8080/products?seller_id=${counter}&xlsx_uri=http://nginx/big_table30k.xlsx"
((counter++))
curl $uri

done
