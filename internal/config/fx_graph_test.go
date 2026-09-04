package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestCoreFxModuleGraphValid(t *testing.T) {
	err := fx.ValidateApp(fx.Supply(&AppConfig{}), CoreFxModule)
	require.NoError(t, err)
}
