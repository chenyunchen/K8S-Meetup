autoscale: true
slidenumbers: true
theme: Ostrich, 5

[.slidenumbers: false]
[.footer: https://github.com/chenyunchen/K8S-Meetup]
![](kubernetes-thumb.png)

## Kubernetes Library 
## 開發實戰
### with client-go
#### **SDN x Cloud Native Meetup #6**

---

[.slidenumbers: false]

## Who am I

Chen Yun Chen (Alex)

- me@yunchen.tw
- blog.yunchen.tw


Experience

- Software Engineer at 
Linker Networks

![left](yun.jpg)

---

##什麼是client-go?
想像是把kubectl拆成各種工具包
Go語言只要拿你所需要的工具
就可以管理集群上想監控的部分

其它語言 Python Java
<https://github.com/kubernetes-client>

---

## 誰會需要 client-go ?

- Debug
- 產品/專案本身特色即為User提供cluster操作 

- 哪些專案也用client-go:
> ![inline](etcd-glyph-color.png) etcd-Operator 
> ![inline](prometheus_color.png) Prometheus-Operator

---

![left](k8s-tree.png)

**api** : Resources Object
**apimachinery** : Options Object

- k8s.io/kubernetes
- k8s.io/client-go
- k8s.io/apiserver

<https://github.com/kubernetes>

---

## api/core/v1/types.go

#### <https://github.com/kubernetes/api/blob/master/core/v1/types.go>

```go
type Volume struct {}
type PersistentVolume struct {}
type PersistentVolumeClaim struct {}
type Container struct {}
type Pod struct {}
type Service struct {}
type Node struct {}
type Namespace struct {}
type ConfigMap struct {}
...
```

---

## apimachinery/pkg/apis/meta/v1/types.go

#### <https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/types.go>

```go
type ListOptions struct {}
type GetOptions struct {}
type DeleteOptions struct {}
type ExportOptions struct {}
...
```

---

## apimachinery/pkg/apis/meta/v1/meta.go

#### <https://github.com/kubernetes/apimachinery/blob/master/pkg/apis/meta/v1/meta.go>

```go
type Object interface {
    GetNamespace() string
    SetNamespace(namespace string)
    GetName() string
    SetName(name string)
    GetGenerateName() string
    SetGenerateName(name string)
    GetUID() types.UID
    SetUID(uid types.UID)
    GetLabels() map[string]string
    SetLabels(labels map[string]string)
    ...
}
```

---

![inline](client-go.png)

---

## Clients Component Type
- Clientset `return struct{}`

> client-go/kubernetes

---

## Clientset

```go
type Clientset struct {
    *discovery.DiscoveryClient
    admissionregistrationV1alpha1 *admissionregistrationv1alpha1.AdmissionregistrationV1alpha1Client
    appsV1                        *appsv1.AppsV1Client
    authenticationV1              *authenticationv1.AuthenticationV1Client
    autoscalingV2beta1            *autoscalingv2beta1.AutoscalingV2beta1Client
    batchV1                       *batchv1.BatchV1Client
    batchV1beta1                  *batchv1beta1.BatchV1beta1Client
    certificatesV1beta1           *certificatesv1beta1.CertificatesV1beta1Client
    coreV1                        *corev1.CoreV1Client
    eventsV1beta1                 *eventsv1beta1.EventsV1beta1Client
    extensionsV1beta1             *extensionsv1beta1.ExtensionsV1beta1Client
    networkingV1                  *networkingv1.NetworkingV1Client
    policyV1beta1                 *policyv1beta1.PolicyV1beta1Client
    schedulingV1alpha1            *schedulingv1alpha1.SchedulingV1alpha1Client
    settingsV1alpha1              *settingsv1alpha1.SettingsV1alpha1Client
    storageV1                     *storagev1.StorageV1Client
    ...
}
```

---

![inline](k8s-clientset-structure.png)

---

![inline](k8s-clientset-deployment.png)

---

## clients 從K8S集群外來管理

```go
import (
    "os"
    "path/filepath"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

...

// Windows: c:\Users\$UserName
os.Getenv("USERPROFILE")

// Linux: /Users/UserName/
os.Getenv("HOME")

// 直接使用kubectl的config(/.kube/config)來產生clientsets
kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
k8s, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
clientset, err := kubernetes.NewForConfig(k8s)
```

---

## clients 從K8S集群內來管理

```go
import (
    "k8s.io/client-go/rest"
    "k8s.io/client-go/kubernetes"
)

...

k8s, err := rest.InClusterConfig()
clientset, err := kubernetes.NewForConfig(k8s)
```

##### **"system:serviceaccount:default:default" cannot get at the cluster scope.**

---

## if 從集群內 : ClusterRole

> 需事先定義要哪些資源的權限

```yaml
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default
  name: resources-editor
rules:
- apiGroups: ["extensions", "apps"]
  resources: ["deployments", "replicasets", "pods", "services", "endpoints", "jobs", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

---

## if 從集群內 : ClusterRoleBinding

> 將定義好的權限賦予至ServiceAccount

```yaml
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: editor-resources
  namespace: default
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
roleRef:
  kind: ClusterRole
  name: resources-editor
  apiGroup: rbac.authorization.k8s.io
```

---

## Discovery

```go
import (
    "fmt"
)

...

version, _ := clientset.Discovery().ServerVersion()
fmt.Println(version)
// v1.11.0

apiList, _ := clientset.Discovery().ServerGroups()
for _, api := range apiList.Groups {
    fmt.Printf("%s : %s \n", api.Name, api.Versions[0].Version)
}
// scheduling.k8s.io : v1beta1
// storage.k8s.io : v1
// networking.k8s.io : v1

resourceList, _ := clientset.Discovery().ServerResources()
for _, r := range resourceList {
    fmt.Printf("%s : %s \n", r.GroupVersion, r.APIResources[0].Name)
}
// apps/v1 : controllerrevisions
// batch/v1 : jobs
// networking.k8s.io/v1 : networkpolicies
```

---

## GET(name, *v1.GetOptions)

> kubectl get pod podname

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

...

clientset.CoreV1().Pods("namespace").Get("name", metav1.GetOptions{})
```

```go
// GetOptions is the standard query options to the standard REST get call.
type GetOptions struct {
    // When specified:
    // - if unset, then the result is returned from remote storage based on quorum-read flag;
    // - if it's 0, then we simply return what we currently have in cache, no guarantee;
    // - if set to non zero, then the result is at least as fresh as given rv.
    ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,1,opt,name=resourceVersion"`
}
```

---

## Create(resource *v1.Resource)

```go
import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

...
containers := []corev1.Container{
    {
        Name:         "busybox",
        Image:        "busybox",
        Command:      []string{"sleep", "3600k"},
    },
}
pod := corev1.Pod{
    ObjectMeta: metav1.ObjectMeta{
        Name: "name",
    },
    Spec: corev1.PodSpec{
        Containers: containers,
    },
}

clientset.CoreV1().Pods("namespace").Create(&pod)
```

---

## List(*v1.ListOptions)

> kubectl get pods

```go
import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

...

pods := []*corev1.Pod{}
podsList, err := clientset.CoreV1().Pods("namespace").List(metav1.ListOptions{})
//podsList.Items: []Pod{}
for _, p := range podsList.Items {
    pods = append(pods, &p)
}
```

```go
// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
    // When specified for list:
    // - if unset, then the result is returned from remote storage based on quorum-read flag;
    // - if it's 0, then we simply return what we currently have in cache, no guarantee;
    // - if set to non zero, then the result is at least as fresh as given rv.
    // +optional
    ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,4,opt,name=resourceVersion"`
}
```

---

## Watch(*v1.ListOptions)

> kubectl get pods --watch

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

...

clientset.CoreV1().Pods("namespace").Watch(metav1.ListOptions{ResourceVersion: "0"})
```

```go
// ListOptions is the query options to a standard REST list call.
type ListOptions struct {
    // When specified for list:
    // - if unset, then the result is returned from remote storage based on quorum-read flag;
    // - if it's 0, then we simply return what we currently have in cache, no guarantee;
    // - if set to non zero, then the result is at least as fresh as given rv.
    // +optional
    ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,4,opt,name=resourceVersion"`
}
```

^ 
ResourceVersion 意義不一樣 是從哪個時間點來將資料吐回
沒設就是隨機 建議設至"0"也能保證資料連續
另外一點是會timeout 10~15min 就算有event也一樣

---

## Update(*v1.Resource)
## UpdateStatus(*v1.Resource)

```go
import (
    "fmt"

    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
)

...

pod := corev1.Pod{}

for {
    _, err := clientset.CoreV1().Pods("namespace").Update(&pod)
    if errors.IsConflict(err) {
        fmt.Println("encountered conflict, retrying")
    } else if err != nil {
        panic(err)
    } else {
        break
    }
}
```

^ Optimistic concurrency via compare and swap fail loop do it!

---

## Patch(name, types.PatchType, []byte)

```go
import (
    corev1 "k8s.io/api/core/v1"
)

...

pod := corev1.Pod{}
patchBytes, err := json.Marshal(pod)
clientset.CoreV1().Pods("namespace").Patch("name", types.JSONPatchType, patchBytes)
```

```go
// Similarly to above, these are constants to support HTTP PATCH utilized by
// both the client and server that didn't make sense for a whole package to be
// dedicated to.
type PatchType string

const (
    JSONPatchType           PatchType = "application/json-patch+json"
    MergePatchType          PatchType = "application/merge-patch+json"
    StrategicMergePatchType PatchType = "application/strategic-merge-patch+json"
)
```

^ 
較update安全
不用像update用loop，自動retry 5 times
建議加上uid 

---

## Delete(name, *v1.DeleteOptions)

> kubectl delete pod podname

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

...

clientset.CoreV1().Pods("namespace").Delete("name", metav1.DeleteOptions{})
```

```go
// DeleteOptions may be provided when deleting an API object.
type DeleteOptions struct {
    // Must be fulfilled before a deletion is carried out. If not possible, a 409 Conflict status will be
    // returned.
    // +optional
    Preconditions *Preconditions `json:"preconditions,omitempty" protobuf:"bytes,2,opt,name=preconditions"`

    // Should the dependent objects be orphaned. If true/false, the "orphan"
    // finalizer will be added to/removed from the object's finalizers list.
    // Either this field or PropagationPolicy may be set, but not both.
    // +optional
    OrphanDependents *bool `json:"orphanDependents,omitempty" protobuf:"varint,3,opt,name=orphanDependents"`
    
	PropagationPolicy *DeletionPropagation `json:"propagationPolicy,omitempty" protobuf:"varint,4,opt,name=propagationPolicy"`
}
```

^
建議加上uid  
DeleteOptions.Preconditions : UID

---

## *v1.DeleteOptions.PropagationPolicy

```go
// Orphans the dependents.
DeletePropagationOrphan DeletionPropagation = "Orphan"
// Deletes the object from the key-value store, the garbage collector will
// delete the dependents in the background.
DeletePropagationBackground DeletionPropagation = "Background"
// The object exists in the key-value store until the garbage collector
// deletes all the dependents whose ownerReference.blockOwnerDeletion=true
// from the key-value store.  API sever will put the "foregroundDeletion"
// finalizer on the object, and sets its deletionTimestamp.  This policy is
// cascading, i.e., the dependents will be deleted with Foreground.
DeletePropagationForeground DeletionPropagation = "Foreground"
```

---

[.slidenumbers: false]
![](level1.jpg)

# **Level 1**
### Create, Update & Delete Deployment
### (More: Restful API server)

---

## (More:)gorilla/mux

```go
import (
    "net/http"
    "log"
    "github.com/gorilla/mux"
)

...

func CreateDeploymentHandler(w http.ResponseWriter, r *http.Request) {
    //Do something here
    w.Write([]byte("Success!\n"))
}

func UpdateDeploymentHandler(w http.ResponseWriter, r *http.Request) {
    //Do something here
    w.Write([]byte("Success!\n"))
}

func DeleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
    //Do something here
    w.Write([]byte("Success!\n"))
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/deployment/create", CreateDeploymentHandler).Methods("POST")
    r.HandleFunc("/deployment/update", UpdateDeploymentHandler).Methods("PUT")
    r.HandleFunc("/deployment/delete", DeleteDeploymentHandler).Methods("Delete")

    log.Fatal(http.ListenAndServe(":8000", r))
}
```

---

## Testing: Fake Clientset

```go
import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes/fake"
)

...

namespace := "default"
pod := corev1.Pod{
  ObjectMeta: metav1.ObjectMeta{
    Name: "K8S-Pod-1",
  },
}
fakeClientset := fake.NewSimpleClientset()
_, err := fakeClientset.CoreV1().Pods(namespace).Create(&pod)
result, err := suite.kubectl.GetPod("K8S-Pod-1")
```

---

## kubernetes.Interface

```go
import (
    "k8s.io/client-go/kubernetes"
)

...

func CreatePod(clientset kubernetes.Interface, namespace string, pod *corev1.Pod) (*corev1.Pod, error) {
    return clientset.CoreV1().Pods(namespace).Create(pod)
}

func GetPod(clientset kubernetes.Interface, namespace string, podName string) (*corev1.Pod, error) {
    return clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
}

func DeletePod(clientset kubernetes.Interface, namespace string, podName string) (*corev1.Pod, error) {
    return clientset.CoreV1().Pods(namespace).Delete(podName, &metav1.GetOptions{})
}
```

---

[.slidenumbers: false]
![](level2.png)

# **Level 2**
### Create, Update & Delete Deployment with Testing

---

## Clients Component Type
- Clientset `return struct{}`

> client-go/kubernetes

- Dynamic Client `return map[string]interface{}`

> client-go/dynamic

---

## Dynamic Client

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    schema "k8s.io/apimachinery/pkg/runtime/schema"
)

...

resource  := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
dynamicClient.Resource(resource, "namespace").List(metav1.ListOptions{})
```

^
Discovery api 版本之後只需要用dynamic就可以操作全部(resource 包含thirdparty)
Only support json serialized

---

## Clients Component Type
- Clientset `return struct{}`

> client-go/kubernetes

- Dynamic Client `return map[string]interface{}`

> client-go/dynamic

- Rest Client `return struct or byte[]`

> client-go/rest


^
clientset適合用來控制k8s原生資源。
如果你需要使用 ThirdPartyResource 就用 Dynamic Client 或 RESTClient。

---

## Rest Client

```go
import (
    corev1 "k8s.io/api/core/v1"
)

...

var pod corev1.Pod
kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
k8s, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
restclient, err := rest.RESTClientFor(k8s)
restClient.Get().Resource("pods").Namespace("namespace").DO().Into(&pod)
restClient.Get().Resource("pods").Namespace("namespace").DORaw() // byte[]
```

^ 
Dynamic clientset and clientset base on  Rest client 
Rest client allow json and protobuf and get all resource

---

![fit](controller.png)

---

## Workqueue

- Delay Queue
延遲某段時間才將資料放入隊列之中 

> (避免Hot-Loop)

- Rate Limitting Queue
控制資料放入隊列之中的速率 

> (基於Delay Queue的實作)

^
確保同一個pod如果多次放入，最後還是只有一個會在queue裡面
Support prometheus monitoring(確認queue放入的速度是否跟得上取出的速度)

---

## Rate Limitting Queue

```go
import (
    "k8s.io/client-go/util/workqueue"
)

...

queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
defer queue.ShutDown()
```

```go
func NewItemExponentialFailureRateLimiter(baseDelay time.Duration, maxDelay time.Duration) RateLimiter {
    return &ItemExponentialFailureRateLimiter{
        failures:  map[interface{}]int{},
        baseDelay: baseDelay,
        maxDelay:  maxDelay,
    }
}

func DefaultItemBasedRateLimiter() RateLimiter {
    return NewItemExponentialFailureRateLimiter(time.Millisecond, 1000*time.Second)
}
```

---

![left fit](queue.png)

**Informer**

- queue.Add(key)

**Worker**

- queue.Get()
- Processing(key)
- queue.AddRateLimited(key)
- queue.Forget(key)
- queue.Done(key)

---

## Informer

```go
import (
    "k8s.io/apimachinery/pkg/fields"
    "k8s.io/client-go/tools/cache"
    corev1 "k8s.io/api/core/v1"
)

...

podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", "default", fields.Everything())

indexer, informer := cache.NewIndexerInformer(podListWatcher, &corev1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
    AddFunc: func(obj interface{}) {
        key, err := cache.MetaNamespaceKeyFunc(obj)
        if err == nil {
            queue.Add(key)
        }
    },
    UpdateFunc: func(old interface{}, new interface{}) {
        key, err := cache.MetaNamespaceKeyFunc(new)
        if err == nil {
            queue.Add(key)
        }
    },
    DeleteFunc: func(obj interface{}) {
        // IndexerInformer uses a delta queue, therefore for deletes we have to use this
        // key function.
        key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
        if err == nil {
            queue.Add(key)
        }
    },
}, cache.Indexers{})
stop := make(chan struct{})
defer close(stop)
go informer.Run(stop)
<-stop
```

^
有快取
會自動幫你監控資料，timeout也會自動重新監控，甚至斷網還會從斷的時間點繼續監控

---

## Worker

```go
import (
    "fmt"
    "k8s.io/client-go/tools/cache"
    "k8s.io/apimachinery/pkg/util/wait"
)

...

if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
    return
}
go wait.Until(runWorker, time.Second, stop)

func runWorker {
    for {
        key, quit := c.queue.Get()
        if quit {
            break
        }
    }
        defer queue.Done(key)
        obj, exists, _ := indexer.GetByKey(key)
        if exists {
            fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*v1.Pod).GetName())
        }
  }
}

```

---

[.slidenumbers: false]
![](level3.jpg)

# **Level 3**
### Create a pod's service detector

---

[.slidenumbers: false]
![left](end.png)

# Q & A
