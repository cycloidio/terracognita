package interpolator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Interpolator is a helper to interpolate values into attributes, by calling Interpolate
// it'll try to find which is the best interpolation, if any.
type Interpolator struct {
	// provider is the provider prefix used for the specific provider that the interpolator
	// is gonna be used on
	provider string
	// resources is a map with all the resources available to interpolate
	// the key is the 'aws_instance.front' (KEY+NAME) and the value
	// is all the attributes it has available to be referenced
	resources map[string]map[string]map[string]string

	// values holds all the possible values on the resources, the
	// key is the value itself "123" and the value is the reference
	// to the resource attribute that has it "${aws_instance.front.id}"
	// used in case o fallback to check if any attribute has the requested value
	values map[string]string
}

// New returns a new intrepolator, expects the provider prefix
func New(provider string) *Interpolator {
	return &Interpolator{
		provider:  provider,
		resources: make(map[string]map[string]map[string]string),
		values:    make(map[string]string),
	}
}

// AddResourceAttributes adds the resource 'r' (aws_instance.front) with the attributes 'a'
// to the internal list of resources. If the resource 'r' already exists it'll be replaced
// with the new set of 'a'
func (i *Interpolator) AddResourceAttributes(r string, a map[string]string) {
	sr := strings.Split(r, ".")
	// This remove the provider from the resource as the reference will never have it
	// so from 'aws_instance' we transform to 'instance'
	sr[0] = strings.Join(strings.Split(sr[0], "_")[1:], "_")
	if _, ok := i.resources[sr[0]]; !ok {
		i.resources[sr[0]] = make(map[string]map[string]string)
	}
	if _, ok := i.resources[sr[0]][sr[1]]; !ok {
		i.resources[sr[0]][sr[1]] = make(map[string]string)
	}
	i.resources[sr[0]][sr[1]] = a
	for k, v := range a {
		i.values[v] = fmt.Sprintf("${%s.%s}", r, k)
	}
}

// Interpolate will try to return the best interpolation for the attribute 'k' with the value 'v'
// by trying to match the 'k' with any resource, like 'virtual_machine_id' with a 'virtual_machine' resource
// which has an attribute with the value 'v'. If none if found the it'll default to check the internal list
// of all possible 'values'. If nothing is found then it'll return '"", false'
func (i *Interpolator) Interpolate(k, v string) (string, bool) {
	sk := strings.Split(k, "_")
	// We generate a ngram of the k separated by '_', meaning 'virtual_machine_id' will
	// be ['virtual_machine_id', 'virtual_machine', 'virtual'] this way we try to find
	// if there is a resource matching from the most possible to the least possible
	ngramk := make([]string, len(sk), len(sk))
	for i := range sk {
		ngramk[len(sk)-(i+1)] = strings.Join(sk[0:i+1], "_")
	}
	for ngi, ng := range ngramk {
		if rns, ok := i.resources[ng]; ok {
			// We try first to find the exact match
			at := i.checkAttributes(sk, v, ngi, ng, rns)
			if at != "" {
				return at, true
			}
		}
		// If no exact match then we try similar match by using regexp and from all the matching
		// resources which one is closer. By closer we just sort by length inversed (shortest first)
		// resources so we know that those have less "differences" from the main key
		matches := make([]string, 0, 0)
		for rk := range i.resources {
			if regexp.MustCompile(fmt.Sprintf("^(.*_)?%s(_.*)?$", ng)).MatchString(rk) {
				matches = append(matches, rk)
			}
		}
		sort.Slice(matches, func(i, j int) bool {
			return len(matches[i]) < len(matches[j])
		})
		for _, rk := range matches {
			rns := i.resources[rk]
			// We try first to find the exact match
			at := i.checkAttributes(sk, v, ngi, rk, rns)
			if at != "" {
				return at, true
			}
		}
	}
	// If we could not find any precise value then we default to the value list
	if ref, ok := i.values[v]; ok {
		return ref, true
	}
	return "", false
}

// checkAttributes will try to find if from the actual list of attribute of the resource there is one matching the expected value.
// To also be more precise, when iterating over the ngram we try to guess the attribute name by inversing the ngram, so for example
// if we try to intrapolate 'virtual_machine_id' which will generate ['virtual_machine_id', 'virtual_machine', 'virtual'] once we
// are on the 'virtual_machine' we try to find the attribute by seeking what's missing on it, in this case the 'id', so we try to
// match it wit the attribute `.id`
func (i *Interpolator) checkAttributes(sk []string, v string, ngi int, ng string, rns map[string]map[string]string) string {
	for rn, attrs := range rns {
		att := strings.Join(sk[(len(sk)-(ngi)):len(sk)], "_")
		if av, ok := attrs[att]; ok && av == v {
			return fmt.Sprintf("${%s.%s.%s}", fmt.Sprintf("%s_%s", i.provider, ng), rn, att)
		}
	}
	// Then if no exact we try to find first one with the same value on the resource
	for rn, attrs := range rns {
		for ak, av := range attrs {
			if av == v {
				return fmt.Sprintf("${%s.%s.%s}", fmt.Sprintf("%s_%s", i.provider, ng), rn, ak)
			}
		}
	}
	return ""
}
