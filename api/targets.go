package api

import (
	"github.com/starkandwayne/shield/db"
)

type TargetFilter struct {
	Plugin string
	Unused YesNo
}

func FetchTargetsList(plugin, unused string) (*[]db.AnnotatedTarget, error) {
	return GetTargets(TargetFilter{
		Plugin: plugin,
		Unused: MaybeString(unused),
	})
}

func GetTargets(filter TargetFilter) (*[]db.AnnotatedTarget, error) {
	uri := ShieldURI("/v1/targets")
	uri.MaybeAddParameter("plugin", filter.Plugin)
	uri.MaybeAddParameter("unused", filter.Unused)

	data := &[]db.AnnotatedTarget{}
	return data, uri.Get(&data)
}

func GetTarget(uuid string) (*db.AnnotatedTarget, error) {
	data := &db.AnnotatedTarget{}
	return data, ShieldURI("/v1/targets/%s", uuid).Get(&data)
}

func CreateTarget(contentJSON string) (*db.AnnotatedTarget, error) {
	data := struct {
		UUID string `json:"uuid"`
	}{}
	err := ShieldURI("/v1/targets").Post(&data, contentJSON)
	if err == nil {
		return GetTarget(data.UUID)
	}
	return nil, err
}

func UpdateTarget(uuid string, contentJSON string) (*db.AnnotatedTarget, error) {
	err := ShieldURI("/v1/target/%s", uuid).Put(nil, contentJSON)
	if err == nil {
		return GetTarget(uuid)
	}
	return nil, err
}

func DeleteTarget(uuid string) error {
	return ShieldURI("/v1/target/%s", uuid).Delete(nil)
}