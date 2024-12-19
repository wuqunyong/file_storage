# 生成PB文件
- protoc.exe --plugin=protoc-gen-go=%BIN_DIR%\protoc-gen-go.exe --go_out=. .\protobuf\pbtest.proto
- protoc.exe --plugin=protoc-gen-doc=%BIN_DIR%\protoc-gen-doc.exe --doc_out=.\protobuf .\protobuf\pbtest.proto












E:\VCity\city\cherry\tools\bin\protoc\bin\protoc.exe -I=E:\VCity\file_storage\protobuf\ --plugin=protoc-gen-go=E:\VCity\city\cherry\tools\bin\pbplugin\protoc-gen-go.exe  --go_out=. rpc_msg.proto

E:\VCity\city\cherry\tools\bin\protoc\bin\protoc.exe -I=E:\VCity\file_storage\protobuf\ --plugin=protoc-gen-go-vtproto=E:\VCity\city\cherry\tools\bin\pbplugin\protoc-gen-go-vtproto.exe  --go-vtproto_out=. rpc_msg.proto