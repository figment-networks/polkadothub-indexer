package timescaleclient

import (
	"github.com/jinzhu/gorm"
	"github.com/figment-networks/polkadothub-indexer/types"
	"reflect"
)

func registerPlugins(c *gorm.DB) {
	c.Callback().Create().Before("gorm:create").Register("db_plugin:before_create", castQuantity)
	c.Callback().Update().Before("gorm:update").Register("db_plugin:before_update", castQuantity)
}

func castQuantity(scope *gorm.Scope) {
	for _, f := range scope.Fields() {
		v := f.Field.Type().String()
		if v == "types.Quantity" {
			f.IsNormal = true
			t := f.Field.Interface().(types.Quantity)
			f.Field = reflect.ValueOf(gorm.Expr("cast(? AS DECIMAL(65,0))", t.String()))
		}
	}
}