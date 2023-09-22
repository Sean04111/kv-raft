# Diff Details

Date : 2023-09-20 23:45:18

Directory d:\\GOPATH__MY\\src\\kv-raft

Total : 59 files,  7706 codes, 1224 comments, 1195 blanks, all 10125 lines

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [.idea/GitLink.xml](/.idea/GitLink.xml) | XML | 6 | 0 | 0 | 6 |
| [.idea/codestream.xml](/.idea/codestream.xml) | XML | 6 | 0 | 0 | 6 |
| [.idea/inspectionProfiles/Project_Default.xml](/.idea/inspectionProfiles/Project_Default.xml) | XML | 27 | 0 | 0 | 27 |
| [.idea/kv-raft.iml](/.idea/kv-raft.iml) | XML | 9 | 0 | 0 | 9 |
| [.idea/modules.xml](/.idea/modules.xml) | XML | 8 | 0 | 0 | 8 |
| [.idea/vcs.xml](/.idea/vcs.xml) | XML | 6 | 0 | 0 | 6 |
| [README.md](/README.md) | Markdown | 3 | 0 | 0 | 3 |
| [go.mod](/go.mod) | Go Module File | 4 | 0 | 4 | 8 |
| [go.sum](/go.sum) | Go Checksum File | 16 | 0 | 1 | 17 |
| [kvraft/client.go](/kvraft/client.go) | Go | 60 | 20 | 11 | 91 |
| [kvraft/common.go](/kvraft/common.go) | Go | 56 | 6 | 13 | 75 |
| [kvraft/config.go](/kvraft/config.go) | Go | 316 | 54 | 56 | 426 |
| [kvraft/kvengine/BST.go](/kvraft/kvengine/BST.go) | Go | 134 | 18 | 9 | 161 |
| [kvraft/kvengine/BST_test.go](/kvraft/kvengine/BST_test.go) | Go | 42 | 0 | 4 | 46 |
| [kvraft/kvengine/checker.go](/kvraft/kvengine/checker.go) | Go | 18 | 4 | 3 | 25 |
| [kvraft/kvengine/compaction.go](/kvraft/kvengine/compaction.go) | Go | 77 | 12 | 17 | 106 |
| [kvraft/kvengine/config.go](/kvraft/kvengine/config.go) | Go | 19 | 8 | 7 | 34 |
| [kvraft/kvengine/database.go](/kvraft/kvengine/database.go) | Go | 7 | 1 | 2 | 10 |
| [kvraft/kvengine/dbfile.go](/kvraft/kvengine/dbfile.go) | Go | 48 | 5 | 6 | 59 |
| [kvraft/kvengine/memtable.go](/kvraft/kvengine/memtable.go) | Go | 7 | 2 | 3 | 12 |
| [kvraft/kvengine/search.go](/kvraft/kvengine/search.go) | Go | 50 | 7 | 9 | 66 |
| [kvraft/kvengine/sstable.go](/kvraft/kvengine/sstable.go) | Go | 229 | 37 | 39 | 305 |
| [kvraft/kvengine/util.go](/kvraft/kvengine/util.go) | Go | 15 | 3 | 6 | 24 |
| [kvraft/kvengine/value.go](/kvraft/kvengine/value.go) | Go | 6 | 0 | 2 | 8 |
| [kvraft/kvengine/wal.go](/kvraft/kvengine/wal.go) | Go | 98 | 12 | 15 | 125 |
| [kvraft/kvengine/wal_test.go](/kvraft/kvengine/wal_test.go) | Go | 26 | 0 | 2 | 28 |
| [kvraft/kvstatemachine.go](/kvraft/kvstatemachine.go) | Go | 28 | 3 | 5 | 36 |
| [kvraft/kvstatemachine_test.go](/kvraft/kvstatemachine_test.go) | Go | 17 | 0 | 3 | 20 |
| [kvraft/logs/0.log](/kvraft/logs/0.log) | Log | 121 | 0 | 1 | 122 |
| [kvraft/logs/1.log](/kvraft/logs/1.log) | Log | 163 | 0 | 1 | 164 |
| [kvraft/logs/2.log](/kvraft/logs/2.log) | Log | 126 | 0 | 1 | 127 |
| [kvraft/logs/3.log](/kvraft/logs/3.log) | Log | 480 | 0 | 1 | 481 |
| [kvraft/logs/4.log](/kvraft/logs/4.log) | Log | 120 | 0 | 1 | 121 |
| [kvraft/logs/5.log](/kvraft/logs/5.log) | Log | 131 | 0 | 1 | 132 |
| [kvraft/logs/6.log](/kvraft/logs/6.log) | Log | 122 | 0 | 1 | 123 |
| [kvraft/server.go](/kvraft/server.go) | Go | 176 | 62 | 32 | 270 |
| [kvraft/test_test.go](/kvraft/test_test.go) | Go | 522 | 90 | 104 | 716 |
| [labgob/labgob.go](/labgob/labgob.go) | Go | 138 | 16 | 21 | 175 |
| [labgob/test_test.go](/labgob/test_test.go) | Go | 128 | 17 | 28 | 173 |
| [labrpc/labrpc.go](/labrpc/labrpc.go) | Go | 333 | 110 | 71 | 514 |
| [labrpc/test_test.go](/labrpc/test_test.go) | Go | 444 | 36 | 118 | 598 |
| [porcupine/bitset.go](/porcupine/bitset.go) | Go | 58 | 2 | 13 | 73 |
| [porcupine/checker.go](/porcupine/checker.go) | Go | 335 | 11 | 28 | 374 |
| [porcupine/model.go](/porcupine/model.go) | Go | 50 | 14 | 14 | 78 |
| [porcupine/porcupine.go](/porcupine/porcupine.go) | Go | 24 | 8 | 8 | 40 |
| [porcupine/visualization.go](/porcupine/visualization.go) | Go | 729 | 103 | 66 | 898 |
| [raft/appendentry.go](/raft/appendentry.go) | Go | 210 | 66 | 32 | 308 |
| [raft/config.go](/raft/config.go) | Go | 439 | 86 | 65 | 590 |
| [raft/election.go](/raft/election.go) | Go | 120 | 60 | 27 | 207 |
| [raft/installsnapshot.go](/raft/installsnapshot.go) | Go | 88 | 29 | 26 | 143 |
| [raft/log.go](/raft/log.go) | Go | 50 | 14 | 9 | 73 |
| [raft/logger.go](/raft/logger.go) | Go | 19 | 78 | 5 | 102 |
| [raft/persister.go](/raft/persister.go) | Go | 54 | 10 | 13 | 77 |
| [raft/raft.go](/raft/raft.go) | Go | 245 | 121 | 53 | 419 |
| [raft/test.sh](/raft/test.sh) | Shell Script | 12 | 2 | 4 | 18 |
| [raft/test_test.go](/raft/test_test.go) | Go | 782 | 90 | 216 | 1,088 |
| [raft/tool.go](/raft/tool.go) | Go | 52 | 4 | 5 | 61 |
| [raft/tool_test.go](/raft/tool_test.go) | Go | 76 | 0 | 8 | 84 |
| [raft/util.go](/raft/util.go) | Go | 21 | 3 | 5 | 29 |

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details