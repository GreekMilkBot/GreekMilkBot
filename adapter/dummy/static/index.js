function timeFormat(date) {
    // 把输入转换为 Date 对象
    const inputDate = new Date(date);
    // 获取当前时间
    const now = new Date();
    // 计算时间差（单位：毫秒）
    const diffMs = now - inputDate;
    // 时间单位换算（毫秒转换）
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    // 根据时间差返回不同的格式
    if (diffSec < 60) {
        return `刚刚`;
    } else if (diffMin < 60) {
        return `${diffMin}分钟前`;
    } else if (diffHour < 24) {
        return `${diffHour}小时前`;
    } else if (diffDay < 7) {
        return `${diffDay}天前`;
    } else {
        // 超过7天，返回具体日期和时间
        const year = inputDate.getFullYear();
        const month = String(inputDate.getMonth() + 1).padStart(2, '0');
        const day = String(inputDate.getDate()).padStart(2, '0');
        const hour = String(inputDate.getHours()).padStart(2, '0');
        const minute = String(inputDate.getMinutes()).padStart(2, '0');
        const second = String(inputDate.getSeconds()).padStart(2, '0');
        return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
    }
}


const users = {
    '10000': {
        name: '自己',
        image: 'https://www.gravatar.com/avatar/68b329da9893e34099c7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10001': {
        name: '张三',
        image: 'https://www.gravatar.com/avatar/68b329da9893e3za99c7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10002': {
        name: '李四',
        image: 'https://www.gravatar.com/avatar/68b329da9893e340adc7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10003': {
        name: '王五',
        image: 'https://www.gravatar.com/avatar/68b329da989cz34099c7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10004': {
        name: '赵六',
        image: 'https://www.gravatar.com/avatar/68b32aaa9893e34099c7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10005': {
        name: '刘七',
        image: 'https://www.gravatar.com/avatar/68b329da9893e3409rr7d8ad5cb9c940?s=200&d=identicon&r=PG'
    },
    '10006': {
        name: '林八',
        image: 'https://www.gravatar.com/avatar/68b329da98qze34099c7d8ad5cb9c940?s=200&d=identicon&r=PG'
    }
}
const groups = {
    '30001': {
        name: '色图俱乐部',
        image: 'https://www.gravatar.com/avatar/68b329da989asas099c7d8ad5cb9c940?s=200&d=identicon&r=PG',
        users: [
            '10000', '10001', '10002', '10003', '10004'
        ]
    }
}
const meta = {
    self: "10000",
    sessions: {
        '20001': {
            type: 'private',
            target: '10001',
        },
        '20002': {
            type: 'private',
            target: '10002',
        },
        '20003': {
            type: 'group',
            target: '30001',
        },
        '20004': {
            type: 'private',
            target: '10004',
        }
    }
}

// 聊天数据
let messages = {
    '20001': [
        {
            id: "900001",
            sender: '10001',
            content: {
                message: parseMessage('你好，今天下午开会讨论项目进展')
            },
            time: '2024-01-01 11:30:30',
        },
        {
            id: "900002",
            sender: '10000',
            content: {
                ref:{
                    sid: '20001',
                    mid:'900001',
                },
                message:[
                    ...parseMessage('好的，我准备好了相关资料'),
                    {
                        type: 'image',
                        data:'https://picsum.photos/400/300?random=10'
                    }
                ]
            },
            time: '2024-01-01 12:32:00',
        },
        {
            id: "900003",
            sender: '10001',
            content: {
                message: parseMessage('太好了，我们3点在会议室见')
            },
            time: '2025-06-20 12:35:00',
        }
    ],
    '20002': [
        {
            id: "900004",
            sender: '10002',
            ref:null,
            content: {
                message: parseMessage('你好，项目文档准备好了吗？')
            },
            time: '2025-06-19 15:20:00',
        },
        {
            id: "900005",
            sender: '10000',
            ref:null,
            content: {
                message: parseMessage('还在整理中，稍后发给你')
            },
            time: '2025-06-19 15:30:00',
        },
        {
            id: "900006",
            sender: '10002',
            ref:null,
            content: {
                message: parseMessage('好的，我稍后回复你')
            },
            time: '2025-06-19 16:00:00',
        }
    ],
    '20003': [
        {
            id: "900007",
            sender: '10002',
            content: {
                message: parseMessage('大家看一下新的需求文档')
            },
            time: '2025-06-18 10:00:00',
        },
        {
            id: "900008",
            sender: '10000',
            content: {
                ref:{
                    sid: '20003',
                    mid:'900007',
                },
                message: parseMessage('@10002 我觉得非常的不错啊')
            },
            time: '2025-06-18 10:10:00',
        },
        {
            id: "900009",
            sender: '10003',
            content: {
                message: parseMessage('我觉得我们需要调整一下开发计划')
            },
            time: '2025-06-19 10:15:00',
        },
        {
            id: "9000010",
            sender: '10004',
            content: {
                message: [
                    ...parseMessage('@10000 同意，我稍后提交一个新的计划'),
                    {
                        type: 'image',
                        data:'https://picsum.photos/400/300?random=10'
                    }
                ]

            },
            time: '2025-06-19 10:30:00',
        }
    ],
    '20004': [
        {
            id: "9000011",
            sender: '10004',
            content: {
                message: parseMessage('周末一起去打球吗？')
            },
            time: '2025-06-18 18:00:00',
        },
        {
            id: "9000012",
            sender: '10000',
            content: {
                message: parseMessage('好的，我周六有时间')
            },
            time: '2025-06-18 18:30:00',
        },
        {
            id: "9000013",
            sender: '10004',
            content: {
                message: parseMessage('那我们周六下午3点老地方见')
            },
            time: '2025-06-18 19:00:00',
        }
    ]
};

function selfInfo() {
    return {
        id: meta.self,
        meta: getUserInfo(meta.self)
    }
}

function getSession(id) {
    let session = meta.sessions[id];
    session['id'] = id
    return session
}

function getSessions() {
    let result = []
    for (const [key, value] of Object.entries(meta.sessions)) {
        value['id'] = key
        result.push(value)
    }
    return result
}

function getUserInfo(id) {
    let user = users[id];
    if (user == null) {
        return null
    }
    user['id'] = id;
    return user;
}

function getGroupInfo(id) {
    return groups[id]
}

function getLastMessage(id) {
    return messages[id][messages[id].length - 1];
}

function getMessages(id) {
    let msg = messages[id]
    for (let item of msg) {
        let u = getUserInfo(item.sender)
        item.name = u.name
        item.image = u.image
        item.lastUpdate = timeFormat(item.time)
        item.isSelf = u.id === selfInfo().id
    }
    return messages[id];
}

function pushMessage(id, msg) {
    messages[id].push(msg);
}

function getMessage(sid, mid) {
    const message = messages[sid];
    for (let data of message) {
        if (data.id === mid) {
            return data
        }
    }
    return {}
}

function searchUser(sessionID, matchUser) {
    let result = []
    if (meta.sessions[sessionID].type === 'group') {
        groups[meta.sessions[sessionID].target].users.forEach(user => {
            const uu = getUserInfo(user)
            if (uu.name.toLowerCase().includes(matchUser) || uu.id.includes(matchUser)) {
                result.push(uu)
            }
        })
    }
    return result
}


function parseMessage(text) {
    const result = [];
    let currentText = '';
    let i = 0;

    while (i < text.length) {
        if (text[i] === '@') {
            // 处理当前累积的文本
            if (currentText) {
                result.push({type: 'text', data: currentText});
                currentText = '';
            }

            // 提取@后的ID（连续非空字符）
            let j = i + 1;
            while (j < text.length && text[j] !== ' ') {
                j++;
            }
            const id = text.substring(i + 1, j);

            // 验证ID是否存在
            if (id && getUserInfo(id) != null) {
                result.push({type: 'at', id});
                i = j;
            } else {
                currentText += '@';
                i++;
            }
        } else {
            currentText += text[i];
            i++;
        }
    }
    if (currentText) {
        result.push({type: 'text', data: currentText});
    }
    return result;
}

function plainMessage(messages) {
    let result = ''
    messages.forEach(message => {
        if(message.type === 'text') {
            result += message.data
        }
        if (message.type === 'at') {
            result += '@'+message.id + ' '
        }
        if (message.type === 'image') {
            result += '[图片]'
        }
    })
    return result
}