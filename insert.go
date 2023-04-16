package cube

func (qb *QueryBuilder) Insert(cols ...string) *QueryBuilder {
	if len(cols) == 0 {
		return qb
	}
	for _, v := range cols {
		qb.inserts = append(qb.inserts, parenthesesWrap(v))
	}
	return qb
}

func (qb *QueryBuilder) Into(table string) *QueryBuilder {
	qb.table = table
	return qb
}

func (qb *QueryBuilder) Values(args ...interface{}) *QueryBuilder {
	qb.args = append(qb.args, args...)
	return qb
}
