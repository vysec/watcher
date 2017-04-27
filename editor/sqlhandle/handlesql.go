package sqlhandle

import (
	"github.com/dearcode/crab/cache"
	"github.com/zssky/log"

	"github.com/dearcode/tracker/editor"
	"github.com/dearcode/tracker/meta"
	"github.com/youtube/vitess/go/vt/sqlparser"
)

var (
	rxe *handlesqlEditor
)

type handlesqlEditor struct {
	args *cache.Cache
}

func init() {
	rxe = &handlesqlEditor{
		args: cache.NewCache(3600),
	}
	editor.Register("sqlhandle", rxe)
}

func (r *handlesqlEditor) Handler(msg *meta.Message, m map[string]interface{}) error {

	err := HandleSql(msg)
	if err != nil {
		return err
	}
	return nil
}

func HandleSql(msg *meta.Message) error {
	return sql(msg)

}

// sql 通过语法树分析sql的类型，表明，条件
func sql(msg *meta.Message) error {
	sql := msg.DataMap["sql"].(string)
	tree, err := sqlparser.Parse(sql)
	sqlparser.Walk(func(node sqlparser.SQLNode) (kontinue bool, err error) {
		switch node.(type) {
		case *sqlparser.Select:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Select).From.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Select).Where.Format(whereBuf)
			msg.DataMap["action"] = "Select"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil
		case *sqlparser.Update:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Update).Where.Format(whereBuf)
			msg.DataMap["action"] = "Update"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil
		case *sqlparser.Delete:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Delete).Where.Format(whereBuf)
			msg.DataMap["action"] = "Delete"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil

		case *sqlparser.Insert:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.Insert).Table.Format(tableBuf)
			whereBuf := sqlparser.NewTrackedBuffer(nil)

			msg.DataMap["action"] = "Insert"
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = whereBuf.String()
			return false, nil

		case *sqlparser.DDL:
			tableBuf := sqlparser.NewTrackedBuffer(nil)
			node.(*sqlparser.DDL).Table.Format(tableBuf)

			msg.DataMap["action"] = node.(*sqlparser.DDL).Action
			msg.DataMap["table"] = tableBuf.String()
			msg.DataMap["condition"] = ""
			return false, nil

		case *sqlparser.Other:
			msg.DataMap["action"] = "Other"
			msg.DataMap["table"] = "Other"
			msg.DataMap["condition"] = "Other"
			return false, nil
		default:
			return true, nil

		}
	}, tree)
	log.Infof("%v", msg.DataMap)
	if err != nil {
		log.Errorf("input: %s, err: %v", sql, err)
		return err
	}
	return nil
}
