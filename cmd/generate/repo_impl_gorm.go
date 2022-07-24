package generate

import (
	"context"
	"fmt"

	"github.com/gin-admin/gin-admin-cli/v5/util"
)

func getModelImplGormFileName(dir, name string) string {
	name = util.ToLowerUnderlinedNamer(name)
	fullname := fmt.Sprintf("%s/internal/app/dao/%s/%s.repo.go", dir, name, name)
	return fullname
}

func genModelImplGorm(ctx context.Context, pkgName, dir string, excludeStatus, excludeCreate bool, item TplItem) error {
	data := map[string]interface{}{
		"PkgName":       pkgName,
		"Name":          item.StructName,
		"PluralName":    util.ToPlural(item.StructName),
		"Comment":       item.Comment,
		"UnderLineName": util.ToLowerUnderlinedNamer(item.StructName),
		"IncludeStatus": !excludeStatus,
		"Fields":        item.Fields,
	}

	buf, err := execParseTpl(daoGromRepoTpl, data)
	if err != nil {
		return err
	}

	fullname := getModelImplGormFileName(dir, item.StructName)
	err = createFile(ctx, fullname, buf)
	if err != nil {
		return err
	}

	fmt.Printf("File write success: %s\n", fullname)

	return execGoFmt(fullname)
}

const daoGromRepoTpl = `
package {{.UnderLineName}}

import (
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"

	"{{.PkgName}}/internal/app/dao/util"
	"{{.PkgName}}/internal/app/schema"
	"{{.PkgName}}/pkg/errors"
)

// Injection wire
var {{.Name}}Set = wire.NewSet(wire.Struct(new({{.Name}}Repo), "*"))

type {{.Name}}Repo struct {
	DB *gorm.DB
}

func (a *{{.Name}}Repo) getQueryOption(opts ...schema.{{.Name}}QueryOptions) schema.{{.Name}}QueryOptions {
	var opt schema.{{.Name}}QueryOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	return opt
}

func (a *{{.Name}}Repo) Query(ctx context.Context, params schema.{{.Name}}QueryParam, opts ...schema.{{.Name}}QueryOptions) (*schema.{{.Name}}QueryResult, error) {
	opt := a.getQueryOption(opts...)

	db := Get{{.Name}}DB(ctx, a.DB)

{{range .Fields}}
	{{if .ConditionArray}}
		if v := params.{{fieldToPlural .StructFieldName}}; {{condition .}} {
			db = db.Where("{{fieldToPluralAndLowerUnderlinedName .StructFieldName}} IN (?)", v)
		}
	{{end}}
	{{if .Condition}}
		if v := params.{{.StructFieldName}}; {{condition .}} {
			db = db.Where("{{fieldToLowerUnderlinedName .StructFieldName}} = ?", v)
		}
	{{end}}
	{{if .ConditionLike}}
		if v := params.{{.StructFieldName}}; {{condition .}} {
			db = db.Where("{{fieldToLowerUnderlinedName .StructFieldName}} LIKE ?", "%"+v+"%")
		}
	{{end}}
{{end}}

	// TODO: Your where condition code here...

	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	}

	if len(opt.OrderFields) > 0 {
		db = db.Order(util.ParseOrder(opt.OrderFields))
	}

	var list {{.PluralName}}
	pr, err := util.WrapPageQuery(ctx, db, params.PaginationParam, &list)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	qr := &schema.{{.Name}}QueryResult{
		PageResult: pr,
		Data:       list.ToSchema{{.PluralName}}(),
	}

	return qr, nil
}

func (a *{{.Name}}Repo) Get(ctx context.Context, id uint64, opts ...schema.{{.Name}}QueryOptions) (*schema.{{.Name}}, error) {
	var item {{.Name}}
	ok, err := util.FindOne(ctx, Get{{.Name}}DB(ctx, a.DB).Where("id=?", id), &item)
	if err != nil {
		return nil, errors.WithStack(err)
	} else if !ok {
		return nil, nil
	}

	return item.ToSchema{{.Name}}(), nil
}

func (a *{{.Name}}Repo) Create(ctx context.Context, item schema.{{.Name}}) error {
	eitem := Schema{{.Name}}(item).To{{.Name}}()
	result := Get{{.Name}}DB(ctx, a.DB).Create(eitem)
	return errors.WithStack(result.Error)
}

func (a *{{.Name}}Repo) Update(ctx context.Context, id uint64, item schema.{{.Name}}) error {
	eitem := Schema{{.Name}}(item).To{{.Name}}()
	result := Get{{.Name}}DB(ctx, a.DB).Where("id=?", id).Updates(eitem)
	return errors.WithStack(result.Error)
}

func (a *{{.Name}}Repo) Delete(ctx context.Context, id uint64) error {
	result := Get{{.Name}}DB(ctx, a.DB).Where("id=?", id).Delete({{.Name}}{})
	return errors.WithStack(result.Error)
}

func (a *{{.Name}}Repo) Truncate(ctx context.Context) error {
	result := Get{{.Name}}DB(ctx, a.DB).Where("1=1").Delete({{.Name}}{})
	return errors.WithStack(result.Error)
}

{{if .IncludeStatus}}
func (a *{{.Name}}Repo) UpdateStatus(ctx context.Context, id uint64, status int) error {
	result := Get{{.Name}}DB(ctx, a.DB).Where("id=?", id).Update("status", status)
	return errors.WithStack(result.Error)
}
{{end}}

`
