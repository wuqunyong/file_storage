# 生成PB文件
- protoc.exe --plugin=protoc-gen-go=%BIN_DIR%\protoc-gen-go.exe --go_out=. .\protobuf\pbtest.proto
- protoc.exe --plugin=protoc-gen-doc=%BIN_DIR%\protoc-gen-doc.exe --doc_out=.\protobuf .\protobuf\pbtest.proto












E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=. rpc_msg.proto

E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go-vtproto=E:\ProtobufBin\protoc-gen-go-vtproto.exe  --go-vtproto_out=. rpc_msg.proto


E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=. login_msg.proto


===
 E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=. rpc_msg/rpc_msg.proto
 E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=. nats_msg/nats_msg.proto



E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. rpc_msg/rpc_msg.proto
E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. nats_msg/nats_msg.proto


E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go-vtproto=E:\ProtobufBin\protoc-gen-go-vtproto.exe  --go-vtproto_out=.    rpc_msg/rpc_msg.proto
E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go-vtproto=E:\ProtobufBin\protoc-gen-go-vtproto.exe  --go-vtproto_out=.    nats_msg/nats_msg.proto



E:\ProtobufBin\protoc-go-inject-tag.exe -input=C:/Users/Administrator/Desktop/编程杂记/file_storage/rpc_msg/rpc_msg.pb.go



[easytcp](https://github.com/DarthPestilane/easytcp)