package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

//CleanStatus omit status field
func CleanStatus(m map[string]interface{}) {
	for k := range m {
		unstructured.RemoveNestedField(m, k)
	}
}

//CleanSvc omit clusterIP/clusterIPs fields
func CleanSvc(f *unstructured.Unstructured) error {
	spec, found, err := unstructured.NestedFieldNoCopy(f.Object, "spec")
	if err != nil {
		return err
	}
	if found {
		s := spec.(map[string]interface{})
		for p := range s {
			if p == "clusterIPs" || p == "clusterIP" {
				delete(s, "clusterIPs")
				delete(s, "clusterIP")
			}
		}
	}
	return nil
}

//CommonCleaning omit unnecessary fields
func CommonCleaning(f *unstructured.Unstructured){
	var t metav1.Time
	f.SetUID("")
	f.SetResourceVersion("")
	f.SetCreationTimestamp(t)
	f.SetDeletionTimestamp(nil)
	f.SetSelfLink("")
	f.SetGenerateName("")
	f.SetGeneration(0)
	f.SetAnnotations(nil)
	f.SetManagedFields(nil)
	f.SetOwnerReferences(nil)
	f.SetFinalizers(nil)
}