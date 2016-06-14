# Create and Manage Persistent Volumes
(In the 6.7 Google Compute Engine section)
https://access.redhat.com/documentation/en/red-hat-enterprise-linux-atomic-host/7/getting-started-with-containers/chapter-6-get-started-provisioning-storage-in-kubernetes
http://kubernetes.io/docs/user-guide/persistent-volumes/walkthrough/

## Test out non persistent disk version of app.
...

##   Create and attach new disk   

create disk
```
gcloud compute disks create mydb-disk --size 10
```


```
gcloud compute instances list
```

attach disk
```
gcloud compute instances attach-disk gke-work-high-mem-d404060a-qbbm --disk mydb-disk
```

#### Format and populate the disk
In shell 2
ssh into node with attached disk
```
gcloud compute --project "askcarter-talks" ssh --zone "us-central1-b" "gke-work-high-mem-d404060a-qbbm"
```

get name of new disk by diffing attached checking attached devices.
it won't have a valid partition -- this is fine.
```
sudo fdisk -l
```

wipe and format the disk as NFS
```
sudo mkfs.ext4 -F -E lazy_itable_init=0,lazy_journal_init=0,discard /dev/sdb
```

remove the mount path (any data written here now will get hidden when we mount the disk)
```
sudo rm -rf /mnt/mydb-disk
```
mount the drive
```
sudo -p mkdir /mnt/mydb-disk
`````
`
sudo mount -o discard,defaults /dev/sdb /mnt/mydb-disk
```

make it read/writeable
```
sudo chmod u+rw /mnt/mydb-disk
```

make it so that the drive automounts on restart.  More info [here](https://community.linuxmint.com/tutorial/view/1513).
```
echo '/dev/sdb /mnt/mydb-disk ext4 discard,defaults 1 1' | sudo tee -a /etc/fstab
```

add data to disk (pull from github, store it in correct location)
```
git clone https://github.com/askcarter/spacerep
`````
`
cd spacerep/cmd/dbd/
`````
`
sudo cp test /mnt/mydb-disk/data
```

#### Exercise: Run the app
In shell 1

detach disk (this is only necessary until K8s 1.3 comes out)
```
gcloud compute instances detach-disk gke-work-high-mem-d404060a-qbbm --disk mydb-disk
```

set up pvc and pv
```
kubectl create -f pvc/mydb.yaml -f pv/mydb.yaml
```

modify deployment to schedule pod on labeled node (cloud.google.com/gke-nodepool=high-mem)
```
nano deployment/persistent-db.yaml
```

run and viola!
```
kubectl create -f deployment/persistent-db.yaml
```

```
kubectl create -f service/persistent-db.yaml
```

```
curl “http://<ExternalIP>list?type=cards&user=admin&q=*”
```

