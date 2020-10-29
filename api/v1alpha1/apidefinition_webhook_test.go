package v1alpha1

import (
	"testing"

	"k8s.io/utils/pointer"
)

func TestApiDefinition_Default(t *testing.T) {
	in := ApiDefinition{
		Spec: APIDefinitionSpec{UseStandardAuth: true},
	}
	in.Default()

	if !in.Spec.VersionData.NotVersioned {
		t.Fatal("expected the api to not be versioned")
	}

	if in.Spec.VersionData.DefaultVersion != "Default" {
		t.Fatal("expected default version to be Default")
	}

	if len(in.Spec.VersionData.Versions) == 0 {
		t.Fatal("expected default version to be applied")
	}

	authConf, ok := in.Spec.AuthConfigs["authToken"]
	if !ok {
		t.Fatal("we used standard auth, so the authToken config must be set")
	}

	if authConf.AuthHeaderName != "Authorization" {
		t.Fatal("expected the authConf.AuthHeaderName to be Authorization, Got", authConf.AuthHeaderName)
	}
}

func TestApiDefinition_Default_DoNotTrack(t *testing.T) {
	in := ApiDefinition{
		Spec: APIDefinitionSpec{
			UseStandardAuth: true,
			DoNotTrack:      pointer.BoolPtr(true),
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != true {
		t.Fatalf("expected DoNotTrack to be true by default, got %v", *in.Spec.DoNotTrack)
	}

	in = ApiDefinition{
		Spec: APIDefinitionSpec{
			UseStandardAuth: true,
			DoNotTrack:      nil,
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != true {
		t.Fatalf("expected DoNotTrack to be true as explicitly set, got %v", *in.Spec.DoNotTrack)
	}

	in = ApiDefinition{
		Spec: APIDefinitionSpec{
			UseStandardAuth: true,
			DoNotTrack:      pointer.BoolPtr(false),
		},
	}
	in.Default()

	if *in.Spec.DoNotTrack != false {
		t.Fatalf("expected DoNotTrack to be false as explicitly set, got %v", *in.Spec.DoNotTrack)
	}
}
