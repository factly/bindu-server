package policy

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/factly/bindu-server/action/role"

	"github.com/factly/bindu-server/model"
	"github.com/spf13/viper"

	"github.com/factly/bindu-server/util"
)

// Mapper map policy
func Mapper(ketoPolicy model.KetoPolicy, userMap map[string]model.User) model.Policy {
	permissions := make([]model.Permission, 0)
	for _, resource := range ketoPolicy.Resources {
		var eachRule model.Permission

		resourcesPrefixAll := strings.Split(resource, ":")
		resourcesPrefix := strings.Join(resourcesPrefixAll[1:], ":")
		eachRule.Resource = resourcesPrefixAll[len(resourcesPrefixAll)-1]
		eachRule.Actions = make([]string, 0)

		for _, action := range ketoPolicy.Actions {
			if strings.HasPrefix(action, "actions:"+resourcesPrefix) {
				actionSplitAll := strings.Split(action, ":")
				eachRule.Actions = append(eachRule.Actions, actionSplitAll[len(actionSplitAll)-1])
			}
		}

		permissions = append(permissions, eachRule)
	}

	subjects := make([]interface{}, 0)
	for _, subject := range ketoPolicy.Subjects {
		_, err := strconv.Atoi(subject)
		if err != nil {
			resp, err := util.Request("GET", viper.GetString("keto_url")+"/engines/acp/ory/regex/roles/"+subject, nil)
			if err != nil {
				continue
			}

			var ketoRole model.KetoRole
			defer resp.Body.Close()
			_ = json.NewDecoder(resp.Body).Decode(&ketoRole)
			role := role.Mapper(ketoRole, userMap)

			subjects = append(subjects, role)
		} else {
			val, exists := userMap[subject]
			if exists {
				subjects = append(subjects, val)
			}
		}

	}

	var result model.Policy
	nameAll := strings.Split(ketoPolicy.ID, ":")
	result.Name = nameAll[len(nameAll)-1]
	result.Description = ketoPolicy.Description
	result.Permissions = permissions
	result.Subjects = subjects

	return result
}

// GetPermissions gives permissions from policy for given userID
func GetPermissions(ketoPolicy model.KetoPolicy, userID uint) []model.Permission {
	permissions := make([]model.Permission, 0)
	for _, resource := range ketoPolicy.Resources {
		var eachRule model.Permission

		resourcesPrefixAll := strings.Split(resource, ":")
		resourcesPrefix := strings.Join(resourcesPrefixAll[1:], ":")
		eachRule.Resource = resourcesPrefixAll[len(resourcesPrefixAll)-1]
		eachRule.Actions = make([]string, 0)

		for _, action := range ketoPolicy.Actions {
			if strings.HasPrefix(action, "actions:"+resourcesPrefix) {
				actionSplitAll := strings.Split(action, ":")
				eachRule.Actions = append(eachRule.Actions, actionSplitAll[len(actionSplitAll)-1])
			}
		}

		permissions = append(permissions, eachRule)
	}

	return permissions
}

// GetAllPolicies gives list of all keto policies
func GetAllPolicies() ([]model.KetoPolicy, error) {
	resp, err := util.Request("GET", viper.GetString("keto_url")+"/engines/acp/ory/regex/policies", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var policyList []model.KetoPolicy
	err = json.NewDecoder(resp.Body).Decode(&policyList)
	if err != nil {
		return nil, err
	}
	return policyList, nil
}
