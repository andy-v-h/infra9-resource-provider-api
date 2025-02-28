// Copyright Infratographer, Inc. and/or licensed to Infratographer, Inc. under one
// or more contributor license agreements. Licensed under the Elastic License 2.0;
// you may not use this file except in compliance with the Elastic License 2.0.
//
// Code generated by entc, DO NOT EDIT.

package generated

import (
	"time"

	"go.infratographer.com/resource-provider-api/internal/ent/generated/resourceprovider"
	"go.infratographer.com/resource-provider-api/internal/ent/schema"
	"go.infratographer.com/x/gidx"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	resourceproviderMixin := schema.ResourceProvider{}.Mixin()
	resourceproviderMixinFields0 := resourceproviderMixin[0].Fields()
	_ = resourceproviderMixinFields0
	resourceproviderFields := schema.ResourceProvider{}.Fields()
	_ = resourceproviderFields
	// resourceproviderDescCreatedAt is the schema descriptor for created_at field.
	resourceproviderDescCreatedAt := resourceproviderMixinFields0[0].Descriptor()
	// resourceprovider.DefaultCreatedAt holds the default value on creation for the created_at field.
	resourceprovider.DefaultCreatedAt = resourceproviderDescCreatedAt.Default.(func() time.Time)
	// resourceproviderDescUpdatedAt is the schema descriptor for updated_at field.
	resourceproviderDescUpdatedAt := resourceproviderMixinFields0[1].Descriptor()
	// resourceprovider.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	resourceprovider.DefaultUpdatedAt = resourceproviderDescUpdatedAt.Default.(func() time.Time)
	// resourceprovider.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	resourceprovider.UpdateDefaultUpdatedAt = resourceproviderDescUpdatedAt.UpdateDefault.(func() time.Time)
	// resourceproviderDescName is the schema descriptor for name field.
	resourceproviderDescName := resourceproviderFields[1].Descriptor()
	// resourceprovider.NameValidator is a validator for the "name" field. It is called by the builders before save.
	resourceprovider.NameValidator = resourceproviderDescName.Validators[0].(func(string) error)
	// resourceproviderDescID is the schema descriptor for id field.
	resourceproviderDescID := resourceproviderFields[0].Descriptor()
	// resourceprovider.DefaultID holds the default value on creation for the id field.
	resourceprovider.DefaultID = resourceproviderDescID.Default.(func() gidx.PrefixedID)
}
