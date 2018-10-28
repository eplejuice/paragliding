### This is the file for the clocktrigger function to be deployed on openstack

For now the webhook Url is hardcoded, because i ran out of time to find a better idea.
However i believe it is better to use some sort of .env variable, to prevent missuse of the webhook.
Hardcoding it displays it as plaintext, so in theory anyone can now flood your webhook with traffic
