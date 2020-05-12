/*
Copyright 2020 The Skaffold Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package image

import (
	"strings"
)

type Registry interface {
	// Name returns the string representation of the registry
	String() string

	// Replace replaces the current registry in a given image name to input registry
	Update(reg *Registry) Registry

	// Prefix gives the prefix for replacing the registry
	Prefix() string

	// Postfix gives the postfix for replacing the registry
	Postfix() string
}

type Image interface {
	// Registry returns the registry for a given image
	Registry() Registry

	// Name returns the image name
	String() string

	// Replace updates the Registry for the image to a new Registry and returns the updated Image
	Update(reg Registry) string
}

// RegistryFactory takesn an input string repo and parses it to return the appropriate registry type
func RegistryFactory(repo string) Registry {
	// Default: return generic registry type
	return NewGenericRegistry(repo)
}

// Factory takes an input string image and parses it to return the appropriate image type
func Factory(image string) Image {
	// Separate repo from image name in string
	splitImage := strings.Split(image, "/")
	imageRegistry := NewGenericRegistry(strings.Join(splitImage[:len(splitImage)-1], "/"))
	imageName := splitImage[len(splitImage)-1]

	// Default: return generic image type
	return NewGenericImage(imageRegistry, imageName)
}
