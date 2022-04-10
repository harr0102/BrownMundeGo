function manInTheMiddleAttack() {
    fetch('/targetdevice/maninthemiddleattack', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        })
        .then(data => {
            console.log(data)
        })
        .catch(error => console.error(error))
}

function dongleAttack() {
    const name = document.getElementById("deviceNamexD").value;
    const commands = document.getElementById("commandList").value.split("\n")


    fetch('/targetdevice/attack', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        name: name,
        commands: commands
    })
    })
    .then(data => {
        console.log(data)
    })
    .catch(error => console.error(error))
}

dongleInit.addEventListener('click', function handleClick() {
    commandList.value = 'ATZ\nATH1\nAT SP 6'
});

filter_testBtn.addEventListener('click', function handleClick() {
    commandList.value = 'ATZ\nATH1\nAT SP 6\nAT SH 191\n04 00 00\n02 00 00\n00 00 00\n01 00 00\n03 00 00\n05 00 00\n07 00 00\n08 00 00\n09 01 00\n10 00 01';
  });

kmlBtn.addEventListener('click', function handleClick() {
    commandList.value = 'ATCRA 7C8\nATFCSH 7C0\n3E1\nATE0\n3BA280';
});

vinpidBtn.addEventListener('click', function handleClick() {
    
    commandList.value = 'ATD\nATE0\nAT AT 0\nATS0\nATH0\nATCAF 1\nAT ST 96\nAT SP 7\n09 02';
});
