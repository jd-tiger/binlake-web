function getLength(str) {
  return str.replace(/[^ -~]/g, 'AA').length;
}


function limitMaxLength(str, maxLength) {
  var result = [];
  for (var i = 0; i < maxLength; i++) {
    var char = str[i]
    if (/[^ -~]/.test(char))
      maxLength--;
    result.push(char);
  }
  return result.join('');
}

function onInput() {
  if (getLength(this.value) > this.maxlength) {
    this.value = limitMaxLength(this.value, maxLength);
  }
}

//字数统计
function chEnWordCount(str) {
  //var count = str.replace(/[^\x00-\xff]/g,"**").length;
  return str.replace(/[^ -~]/g, 'AA').length;
  //return count;
}

//英文字母和数字
function checkPath(str) {
  if (!str.match(/^[A-Za-z0-9]{4,40}$/)) {
    return false;
  }
  return true;
}

// 检查域名格式 域名:端口或ip:端口(ip仅限弹性数据库)
function checkDomain(str) {
  if (!str.match(/^.*.jddb.com:[0-9]{2,5}$/) && !str.match(/^(\d|[1-9]\d|1\d{2}|2[0-5][0-5])\.(\d|[1-9]\d|1\d{2}|2[0-5][0-5])\.(\d|[1-9]\d|1\d{2}|2[0-5][0-5])\.(\d|[1-9]\d|1\d{2}|2[0-5][0-5])\:[0-9]{2,5}$/)) {
    return false;
  }
  return true;
}

//字符串非空检查
function checkEmpty(str) {
  if (str.length <= 0) {
    return true;
  }
  return false;
}

// Trim
function Trim(str) {
  return str.replace(/(^\s*)|(\s*$)/g, "");
}

// LTrim
function LTrim(str) {
  return str.replace(/(^\s*)/g, "");
}

// RTrim
function RTrim(str) {
  return str.replace(/(\s*$)/g, "");
}

$(function(){
    console.log(1111);
    $('.collapse-link').on('click', function () {
        var ibox = $(this).closest('div.ibox');
        var button = $(this).find('i');
        var content = ibox.find('div.ibox-content');
        content.slideToggle(200);
        button.toggleClass('fa-chevron-up').toggleClass('fa-chevron-down');
        ibox.toggleClass('').toggleClass('border-bottom');
        setTimeout(function () {
            ibox.resize();
            ibox.find('[id^=map-]').resize();
        }, 50);
    });

    // Close ibox function
    $('.close-link').on('click', function () {
        var content = $(this).closest('div.ibox');
        content.remove();
    });
})

