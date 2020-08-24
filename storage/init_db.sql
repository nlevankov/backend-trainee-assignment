CREATE TABLE "chats_users" ("user_id" integer,"chat_id" integer, PRIMARY KEY ("user_id","chat_id"));
CREATE TABLE "users" ("id" serial,"name" text NOT NULL UNIQUE,"created_at" timestamp with time zone , PRIMARY KEY ("id"));
CREATE TABLE "chats" ("id" serial,"name" text NOT NULL UNIQUE,"created_at" timestamp with time zone , PRIMARY KEY ("id"));
CREATE TABLE "messages" ("id" serial,"chat_id" integer,"user_id" integer,"text" text NOT NULL,"created_at" timestamp with time zone , PRIMARY KEY ("id"));
ALTER TABLE "messages" ADD CONSTRAINT messages_chat_id_chats_id_foreign FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "messages" ADD CONSTRAINT messages_user_id_users_id_foreign FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;