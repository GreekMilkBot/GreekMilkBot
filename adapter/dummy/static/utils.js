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
            result += '@'+message.data + ' '
        }
        if (message.type === 'image') {
            result += '[图片]'
        }
    })
    return result
}

function fetchObj(url){
    try {
        const xhr = new XMLHttpRequest();
        xhr.open('GET', url, false); // 第三个参数设为 false 表示同步请求
        xhr.send();

        if (xhr.status >= 200 && xhr.status < 300) {
            return JSON.parse(xhr.responseText)
        } else {
            console.log(`请求失败，状态码: ${xhr.status}`);
            return {}
        }
    } catch (e) {
        console.log(e);
        return {}
    }
}


function postObj(url,data){
    const xhr = new XMLHttpRequest();
    xhr.open('POST', url, false);
    xhr.setRequestHeader('Content-Type', 'application/json');
    try {
        xhr.send(JSON.stringify(data));
        if (xhr.status >= 200 && xhr.status < 300) {
            return xhr.responseText
        } else {
            console.log(`请求失败，状态码: ${xhr.status}`);
            return {}
        }
    }catch (e){
        console.log(e);
        return {}
    }
}