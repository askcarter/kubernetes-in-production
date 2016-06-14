# Create and manage Horizontal Pod Autoscaling (HPA)
From:  http://kubernetes.io/docs/user-guide/horizontal-pod-autoscaling/walkthrough/

```
kubectl create -f deployment/myapp.yaml
```

```
kubectl create -f service/myapp.yaml
```

```
kubectl create -f hpa/myapp.yaml
```

```
kubectl get pods 
```

## Exercise: Add load to our application 
In another command shell
```
kubectl exec it my-app-<xxxx> -- sh
```

```
# while true; do wget -O- http://127.0.0.1/heavy-workload; done
```

```
kubectl get deployments my-app; kubectl get hpa my-app
```

## Exercise: Kill load in other shell and watch results

```
kubectl get deployments my-app 
```

```
kubectl get hpa my-app
```
