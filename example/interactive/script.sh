for i in {1..5};do
    echo "Hello, who are you?"
    read  varname
    echo "Hi $varname `hostname -f`" 
    sleep 1
done
