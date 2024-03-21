%: %.go
	go build -o $@ $<

step0_repl/step0_repl: step0_repl/step0_repl.go
step1_read_print/step1_read_print: step1_read_print/step1_read_print.go common/printer.go common/reader.go common/types.go
step2_eval/step2_eval: step2_eval/step2_eval.go common/printer.go common/reader.go common/types.go
step3_env/step3_env: step3_env/step3_env.go common/printer.go common/reader.go common/types.go common/env.go
step4_if_fn_do/step4_if_fn_do: step4_if_fn_do/step4_if_fn_do.go common/printer.go common/reader.go common/types.go common/env.go common/core.go
step5_tco/step5_tco: step5_tco/step5_tco.go common/printer.go common/reader.go common/types.go common/env.go common/core.go
step6_file/step6_file: step6_file/step6_file.go common/printer.go common/reader.go common/types.go common/env.go common/core.go
