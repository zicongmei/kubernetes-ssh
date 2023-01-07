# kubernetes-ssh

A sample code to setup pods with ssh.

Usage:

```
NAMESPACE=ns3
go run controller/cmd/main.go  -namespace $NAMESPACE
kubectl exec -it -n $NAMESPACE sample-0-7dd65f9967-shhhv bash 

ssh -p 2222 sample-1
```