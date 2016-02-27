func Do(request Interface{}) {

	v := reflect.ValueOf(request)
	typ := reflect.TypeOf(request)

	params := make(map[string]string, 0) 

	for i := 0; i < typ.NumField(); i++ { // HL

		name := typ.Field(i).Name
		tag := typ.Field(i).Tag.Get("xc")
		field := v.Field(i)

		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num := field.Int()
			if num == 0 && tag == "optional" {
				continue
			}
			params[name] = strconv.FormatInt(num, 10)
		case reflect.String:
			str := field.String()
			params[name] = str
		}
	}
	typ_name_list := strings.Split(typ.String(), ".")
	typ_name := typ_name_list[len(typ_name_list)-1]
	err := u.Get(params, rsp)
	return rsp, err
}
