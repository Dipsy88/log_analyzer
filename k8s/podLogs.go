package main

import (
	"archive/tar"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/deprecated/scheme"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	rest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

var files = [...]string{"/tmp/file.txt", "/tmp/file2.log"}

func main() {
	namespace := flag.String("namespace", "default", "Namespace to get the logs for the pods")
	destPath := flag.String("destination", "output", "Destination to copy the log files")
	authenticationRequired := flag.Bool("needAuth", false, "Do you need to pass password (optional)")
	userName := flag.String("username", "", "Username to login (optional)")
	flag.Parse()

	if *authenticationRequired {
		if *userName == "" {
			flag.Usage()
			os.Exit(1)
		}
		fmt.Print("Enter password: ")
		password, _ := terminal.ReadPassword(int(syscall.Stdin))
		login(*userName, string(password[:]), *namespace)
	}

	makeDirIfRequired(*destPath)

	listOptions := metav1.ListOptions{}
	ctx := context.Background()

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	config, k8sClient, err := getClient(*kubeconfig)
	check(err)

	pods, err := k8sClient.Pods(*namespace).List(ctx, listOptions)
	check(err)

	coreClient, _ := initRestClient(config)

	for _, pod := range pods.Items {
		nameSpace := pod.GetNamespace()
		podName := pod.GetName()
		fmt.Println(podName, nameSpace)

		writePodLog(*destPath+"/"+podName+".log", k8sClient, nameSpace, &podName)

		for _, val := range files {
			copyFromPod(val, *destPath, config, coreClient, nameSpace, podName, pod.GetClusterName())
		}

	}
	fmt.Println("Process finished")
}

func makeDirIfRequired(path string) {
	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		fmt.Println("Creating directory", path)
		err := os.MkdirAll(path, 0755)
		check(err)
	}
}

func login(user string, password string, namespace string) {

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writePodLog(destPath string, k8sClient typev1.CoreV1Interface, namespace string, pod *string) {
	podLogOpts := coreV1.PodLogOptions{}
	req := k8sClient.Pods(namespace).GetLogs(*pod, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	check(err)

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	check(err)

	f, err := os.Create(destPath)
	check(err)

	defer f.Close()

	f.WriteString(buf.String())
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func initRestClient(config *rest.Config) (*typev1.CoreV1Client, error) {
	coreClient, err := typev1.NewForConfig(config)
	check(err)
	return coreClient, err
}

func copyFromPod(srcPath string, srcDest string, config *rest.Config, coreClient *typev1.CoreV1Client,
	nameSpace string, pod string, container string) error {
	reader, outStream := io.Pipe()
	cmdArr := []string{"tar", "cf", "-", srcPath}
	req := coreClient.RESTClient().
		Get().
		Namespace(nameSpace).
		Resource("pods").
		Name(pod).
		SubResource("exec").
		VersionedParams(&coreV1.PodExecOptions{
			Container: container,
			Command:   cmdArr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	check(err)
	go func() {
		defer outStream.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: outStream,
			Stderr: os.Stderr,
			Tty:    false,
		})
	}()
	prefix := getPrefix(srcPath)
	prefix = path.Clean(prefix)
	//prefix = cpStripPathShortcuts(prefix)
	srcDest = path.Join(srcDest, path.Base(prefix))
	err = untarAll(reader, srcDest, prefix)
	return err
}

func untarAll(reader io.Reader, destDir, prefix string) error {
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		if !strings.HasPrefix(header.Name, prefix) {
			return fmt.Errorf("tar contents corrupted")
		}

		mode := header.FileInfo().Mode()
		destFileName := filepath.Join(destDir, header.Name[len(prefix):])

		baseName := filepath.Dir(destFileName)
		if err := os.MkdirAll(baseName, 0755); err != nil {
			return err
		}
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(destFileName, 0755); err != nil {
				return err
			}
			continue
		}

		evaledPath, err := filepath.EvalSymlinks(baseName)
		check(err)

		if mode&os.ModeSymlink != 0 {
			linkname := header.Linkname

			if !filepath.IsAbs(linkname) {
				_ = filepath.Join(evaledPath, linkname)
			}

			if err := os.Symlink(linkname, destFileName); err != nil {
				return err
			}
		} else {
			outFile, err := os.Create(destFileName)
			check(err)
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

func getPrefix(file string) string {
	return strings.TrimLeft(file, "/")
}

func getClient(kubeconfig string) (*rest.Config, typev1.CoreV1Interface, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	check(err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	check(err)

	return config, clientset.CoreV1(), nil
}
