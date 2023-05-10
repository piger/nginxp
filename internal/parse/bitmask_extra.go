package parse

// Here we can add additional directives that are not present in crossplane.

func init() {
	// Opentracing
	// https://github.com/opentracing-contrib/nginx-opentracing/blob/master/doc/Reference.md
	dirMask["opentracing"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_CONF_FLAG}
	dirMask["opentracing_load_tracer"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_CONF_TAKE2}
	dirMask["opentracing_propagate_context"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_CONF_NOARGS}
	dirMask["opentracing_tag"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_CONF_TAKE2}
	dirMask["opentracing_operation_name"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_CONF_TAKE1}
	dirMask["opentracing_trace_locations"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_CONF_FLAG}

	// ngx_http_js_module
	// http://nginx.org/en/docs/http/ngx_http_js_module.html
	dirMask["js_import"] = []int{NGX_HTTP_MAIN_CONF | NGX_CONF_TAKE13}
	dirMask["js_set"] = []int{NGX_HTTP_MAIN_CONF | NGX_CONF_TAKE12}

	// lua
	// https://github.com/openresty/lua-nginx-module
	dirMask["rewrite_by_lua_file"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_HTTP_LIF_CONF | NGX_CONF_TAKE1}
	dirMask["access_by_lua_file"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_HTTP_LIF_CONF | NGX_CONF_TAKE1}
	dirMask["access_by_lua_block"] = []int{NGX_HTTP_MAIN_CONF | NGX_HTTP_SRV_CONF | NGX_HTTP_LOC_CONF | NGX_HTTP_LIF_CONF | NGX_CONF_BLOCK}
}
