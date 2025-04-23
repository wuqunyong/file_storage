@REM # 生成PB文件
@REM - protoc.exe --plugin=protoc-gen-go=%BIN_DIR%\protoc-gen-go.exe --go_out=. .\protobuf\pbtest.proto
@REM - protoc.exe --plugin=protoc-gen-doc=%BIN_DIR%\protoc-gen-doc.exe --doc_out=.\protobuf .\protobuf\pbtest.proto
@REM E:\ProtobufBin\protoc.exe -I=protobuf --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=. login_msg.proto






E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --plugin=protoc-gen-doc=E:\ProtobufBin\protoc-gen-doc.exe --go_out=. --doc_out=.\proto --go_opt=paths=source_relative --proto_path=. ./proto/rpc_msg/rpc_msg.proto
E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --plugin=protoc-gen-doc=E:\ProtobufBin\protoc-gen-doc.exe --go_out=. --doc_out=.\proto --go_opt=paths=source_relative --proto_path=. ./proto/nats_msg/nats_msg.proto


E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. ./proto/common_msg/common.proto
E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. ./proto/common_msg/login.proto
E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. ./proto/common_msg/pbtest.proto
E:\ProtobufBin\protoc.exe --plugin=protoc-gen-go=E:\ProtobufBin\protoc-gen-go.exe  --go_out=.   --go_opt=paths=source_relative --proto_path=. ./proto/common_msg/service_discovery.proto





@REM E:\ProtobufBin\protoc.exe -I=proto --plugin=protoc-gen-go-vtproto=E:\ProtobufBin\protoc-gen-go-vtproto.exe  --go-vtproto_out=.    rpc_msg/rpc_msg.proto
@REM E:\ProtobufBin\protoc.exe -I=proto --plugin=protoc-gen-go-vtproto=E:\ProtobufBin\protoc-gen-go-vtproto.exe  --go-vtproto_out=.    nats_msg/nats_msg.proto

@REM E:\ProtobufBin\protoc-go-inject-tag.exe -input=C:/Users/Administrator/Desktop/编程杂记/file_storage/rpc_msg/rpc_msg.pb.go
@REM E:\ProtobufBin\protoc-go-inject-tag.exe -input=C:/Users/Administrator/Desktop/编程杂记/file_storage/proto/pbtest.pb.go