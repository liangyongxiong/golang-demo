package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"path/filepath"
)

const (
	k8sNamespace  = "lyx-prj-jnrpm7np"
	k8sDeployment = "lyx-deployment"
)

func GetK8sClientset() *kubernetes.Clientset {
	kubeConfigFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

func RegisterNamespaceInformer(stopCh <-chan struct{}, deleteCh, updateCh chan<- int) {
	clientset := GetK8sClientset()
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	resourceCore := informerFactory.Core().V1().Namespaces()
	resourceInformer := resourceCore.Informer()
	resourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//resource := obj.(*corev1.Namespace)
			//if resource.Name == k8sNamespace {
			//	fmt.Printf("create %s\n", resource.Name)
			//}
		},
		DeleteFunc: func(obj interface{}) {
			resource := obj.(*corev1.Namespace)
			if resource.Name == k8sNamespace {
				deleteCh <- 1
				fmt.Printf("delete namespace %s\n", resource.Name)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old_resource := oldObj.(*corev1.Namespace)
			new_resource := newObj.(*corev1.Namespace)
			if new_resource.Name == k8sNamespace {
				fmt.Printf("update namespace from %s:%s to %s:%s\n", old_resource.Name, old_resource.Status.Phase, new_resource.Name, new_resource.Status.Phase)
				if new_resource.Status.Phase == corev1.NamespaceActive {
					updateCh <- 1
				}
			}
		},
	})

	resourceLister := resourceCore.Lister()
	namespaces, err := resourceLister.List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for i, namespace := range namespaces {
		fmt.Printf("namespaces[%d] %s\n", i, namespace.Name)
	}

	informerFactory.Start(stopCh)
	//informerFactory.WaitForCacheSync(stopCh)
}

func RegisterDeploymentInformer(stopCh <-chan struct{}, deleteCh, updateCh chan<- int, replicas *int32) {
	clientset := GetK8sClientset()
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	resourceApps := informerFactory.Apps().V1().Deployments()
	resourceInformer := resourceApps.Informer()
	resourceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//resource := obj.(*appsv1.Deployment)
			//if resource.Name == k8sDeployment {
			//	fmt.Printf("create %s\n", resource.Name)
			//}
		},
		DeleteFunc: func(obj interface{}) {
			resource := obj.(*appsv1.Deployment)
			if resource.Name == k8sDeployment {
				deleteCh <- 1
				fmt.Printf("delete deployment %s\n", resource.Name)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			old_resource := oldObj.(*appsv1.Deployment)
			new_resource := newObj.(*appsv1.Deployment)
			if new_resource.Name == k8sDeployment {
				fmt.Printf(
					"update deployment from %s:[%d/%d] to %s:[%d/%d]\n",
					old_resource.Name, old_resource.Status.ReadyReplicas, old_resource.Status.Replicas,
					new_resource.Name, new_resource.Status.ReadyReplicas, new_resource.Status.Replicas)
				if new_resource.Status.ReadyReplicas == *replicas {
					updateCh <- 1
				}
			}
		},
	})

	resourceLister := resourceApps.Lister()
	namespaces, err := resourceLister.List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for i, namespace := range namespaces {
		fmt.Printf("namespaces[%d] %s\n", i, namespace.Name)
	}

	informerFactory.Start(stopCh)
	//informerFactory.WaitForCacheSync(stopCh)
}

func GetAllK8sNamespaces() *corev1.NamespaceList {
	clientset := GetK8sClientset()
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return namespaces
}

func GetK8sNamespace(namespaceName string) *corev1.Namespace {
	clientset := GetK8sClientset()
	namespace, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		serr, ok := err.(*errors.StatusError)
		if !ok {
			panic(err)
		}
		fmt.Printf("get namespace %s failed\n", namespaceName)
		fmt.Printf("Reason: %s, Message: %s\n", serr.ErrStatus.Reason, serr.ErrStatus.Message)
		if serr.ErrStatus.Reason == "NotFound" {
			return nil
		}
		panic(serr)
	}
	fmt.Printf(
		"name: %s, status: %s, create: %s\n",
		namespace.ObjectMeta.Name, namespace.Status.Phase, namespace.CreationTimestamp.Format("2006-01-02 15:04:05"))
	return namespace
}

func DeleteK8sNamespace(namespaceName string, ch <-chan int) {
	clientset := GetK8sClientset()
	policy := metav1.DeletePropagationForeground
	err := clientset.CoreV1().Namespaces().Delete(context.TODO(), namespaceName, metav1.DeleteOptions{PropagationPolicy: &policy})
	if err != nil {
		panic(err)
	}
	<-ch
	fmt.Printf("delete namespace %s success\n", namespaceName)
}

func CreateK8sNamespace(namespaceName string, ch <-chan int) {
	clientset := GetK8sClientset()
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
	if err != nil {
		serr, ok := err.(*errors.StatusError)
		if !ok {
			panic(err)
		}
		fmt.Printf("create namespace %s failed\n", namespaceName)
		fmt.Printf("Reason: %s, Message: %s\n", serr.ErrStatus.Reason, serr.ErrStatus.Message)
		panic(serr)
	}
	<-ch
	fmt.Printf("create namespace %s success\n", namespaceName)
}

func DeleteAllK8sDeployments(namespaceName string, ch <-chan int) {
	clientset := GetK8sClientset()
	deployments, err := clientset.AppsV1().Deployments(namespaceName).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, deploy := range deployments.Items {
		err := clientset.AppsV1().Deployments(namespaceName).Delete(context.TODO(), deploy.Name, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
		if deploy.Name == k8sDeployment {
			<-ch
		}
	}
	fmt.Printf("delete deployments of namespace %s success\n", namespaceName)
}

func CreateK8sDeployment(namespaceName, deploymentName string, replicas int32, ch <-chan int) {
	clientset := GetK8sClientset()
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
			Labels: map[string]string{
				"app": "nginx",
			},
			Namespace: namespaceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := clientset.AppsV1().Deployments(namespaceName).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	<-ch
	fmt.Printf("create deployment %s of namespace %s success\n", deploymentName, namespaceName)
}

func PatchK8sDeployment(namespaceName, deploymentName string, replicas int32, ch <-chan int) {
	data := []byte(fmt.Sprintf(`{
		"spec": {
			"replicas": %d
		}
	}`, replicas))

	clientset := GetK8sClientset()
	_, err := clientset.AppsV1().Deployments(namespaceName).Patch(
		context.TODO(), k8sDeployment, types.StrategicMergePatchType, data, metav1.PatchOptions{})
	if err != nil {
		panic(err)
	}
	<-ch
	fmt.Printf("patch deployment %s of namespace %s success\n", deploymentName, namespaceName)
}

func GetAllK8sPods(namespaceName, deploymentName string) *corev1.PodList {
	clientset := GetK8sClientset()
	pods, err := clientset.CoreV1().Pods(namespaceName).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("get all pods of namespace %s success\n", namespaceName)
	return pods
}

func main() {
	var replicas int32 = 1

	namespaceStopCh := make(chan struct{})
	defer close(namespaceStopCh)
	namespaceDeleteCh := make(chan int, 1)
	defer close(namespaceDeleteCh)
	namespaceUpdateCh := make(chan int, 1)
	defer close(namespaceUpdateCh)
	go RegisterNamespaceInformer(namespaceStopCh, namespaceDeleteCh, namespaceUpdateCh)

	deploymentStopCh := make(chan struct{})
	defer close(deploymentStopCh)
	deploymentUpdateCh := make(chan int, 1)
	defer close(deploymentUpdateCh)
	deploymentDeleteCh := make(chan int, 1)
	defer close(deploymentDeleteCh)
	go RegisterDeploymentInformer(deploymentStopCh, deploymentDeleteCh, deploymentUpdateCh, &replicas)

	// 获取全部namespace列表
	//namespaces := GetAllK8sNamespaces()
	//for i, ns := range namespaces.Items {
	//	fmt.Printf(
	//		"namespaces[%d]: name=%s, status=%s, create=%s\n",
	//		i, ns.ObjectMeta.Name, ns.Status.Phase, ns.CreationTimestamp.Format("2006-01-02 15:04:05"))
	//}

	// 获取namespace，如果namespace已存在，则先删除
	namespace := GetK8sNamespace(k8sNamespace)
	if namespace != nil {
		// 先删除deployments
		DeleteAllK8sDeployments(k8sNamespace, deploymentDeleteCh)

		// 再删除namespace
		DeleteK8sNamespace(k8sNamespace, namespaceDeleteCh)
	}

	// 创建namespace
	CreateK8sNamespace(k8sNamespace, namespaceUpdateCh)

	// 创建deployment
	CreateK8sDeployment(k8sNamespace, k8sDeployment, replicas, deploymentUpdateCh)

	// 修改deployment
	replicas = 2
	PatchK8sDeployment(k8sNamespace, k8sDeployment, replicas, deploymentUpdateCh)

	// 获取全部pod列表
	pods := GetAllK8sPods(k8sNamespace, k8sDeployment)
	for i, po := range pods.Items {
		fmt.Printf("pods[%d]: name=%s, phase=%s, reaseon=%s\n", i, po.Name, po.Status.Phase, po.Status.Reason)
		fmt.Printf(
			"host_ip=%s, host_port=%d, pod_ip=%s, container_port=%d\n",
			po.Status.HostIP, 0, po.Status.PodIP, po.Spec.Containers[0].Ports[0].ContainerPort)
	}
}
