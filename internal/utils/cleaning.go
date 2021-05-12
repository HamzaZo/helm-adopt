package utils


import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func CleanStatus(m map[string]interface{}) {
	for k := range m {
		unstructured.RemoveNestedField(m, k)
	}
}

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