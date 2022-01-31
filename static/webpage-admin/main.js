socket = new WebSocket("ws://raspberrypi:3000/api");

connectedToPhysicalJabuti = false;

msgType_commandGetDiscreteMovementConfig = 'f';
msgType_commandSetDiscreteMovementConfig = 'g';

msgType_commandMoveForward = 'c';
msgType_commandMoveLeft = 'd';
msgType_commandMoveRight = 'e';

socket.onopen = function(e) {
    physicalJabuti_onConnection();

    console.log("[socket.onopen] Connection to server stablished");

};

socket.onmessage = function(event) {
    console.log("[socket.onmessage]");
    console.log(`Data received from server: ${event.data}`);

    message = event.data
    msgtype = message[0]

    if(msgtype == msgType_commandGetDiscreteMovementConfig) {
        updateInputsForDiscreteMvConfigs(message);
    }



};

socket.onclose = function(event) {
    physicalJabuti_onDisconnection();

    if(event.wasClean) {
        console.log(`[socket.onclose] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        // e.g. server process killed or network down
        // event.code is usually 1006 in this case
        console.log(`[socket.onclose] Connection died, code=${event.code} reason=${event.reason}`);
    }
};

socket.onerror = function(error) {
    physicalJabuti_onDisconnection();

    console.log("[socket.onerror]: " + error.message);
};

function physicalJabuti_onConnection() {
    document.getElementById("connectionStatusLabel").innerHTML = "Conexão com o jabuti físico: Conectado";
    connectedToPhysicalJabuti = true;
    askServerForMovementString();
}

function physicalJabuti_onDisconnection() {
    document.getElementById("connectionStatusLabel").innerHTML = "Conexão com o jabuti físico: Desconectado";
    connectedToPhysicalJabuti = false;
}

function getDiscreteMovementConfigMessageString() {
    messageString = msgType_commandSetDiscreteMovementConfig;
    messageString += '-';

    messageString += document.getElementById('forward-leftWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('forward-rightWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('forward-timeApplyingPower').value;
    messageString += '-';
    messageString += document.getElementById('forward-timeToWaitAfterPowerApplied').value;
    messageString += '-';

    messageString += document.getElementById('left-leftWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('left-rightWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('left-timeApplyingPower').value;
    messageString += '-';
    messageString += document.getElementById('left-timeToWaitAfterPowerApplied').value;
    messageString += '-';

    messageString += document.getElementById('right-leftWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('right-rightWheelPower-percentage').value;
    messageString += '-';
    messageString += document.getElementById('right-timeApplyingPower').value;
    messageString += '-';
    messageString += document.getElementById('right-timeToWaitAfterPowerApplied').value;

    return messageString;
}

function askServerForMovementString() {
    socket.send(msgType_commandGetDiscreteMovementConfig);
}

function saveDiscreteMvConfig_callback() {
    messageString = getDiscreteMovementConfigMessageString();
    console.log("sending: " + messageString);
    socket.send(messageString);
}

function reloadDiscreteMvConfig_callback() {
    askServerForMovementString();
}

function updateInputsForDiscreteMvConfigs(message) {
    tokens = message.split('-');

    document.getElementById('forward-leftWheelPower-percentage').value = parseFloat(tokens[1]);
    document.getElementById('forward-rightWheelPower-percentage').value = parseFloat(tokens[2]);
    document.getElementById('forward-timeApplyingPower').value = parseInt(tokens[3]);
    document.getElementById('forward-timeToWaitAfterPowerApplied').value = parseInt(tokens[4]);

    document.getElementById('left-leftWheelPower-percentage').value = parseFloat(tokens[5]);
    document.getElementById('left-rightWheelPower-percentage').value = parseFloat(tokens[6]);
    document.getElementById('left-timeApplyingPower').value = parseInt(tokens[7]);
    document.getElementById('left-timeToWaitAfterPowerApplied').value = parseInt(tokens[8]);

    document.getElementById('right-leftWheelPower-percentage').value = parseFloat(tokens[9]);
    document.getElementById('right-rightWheelPower-percentage').value = parseFloat(tokens[10]);
    document.getElementById('right-timeApplyingPower').value = parseInt(tokens[11]);
    document.getElementById('right-timeToWaitAfterPowerApplied').value = parseInt(tokens[12]);
}

function testForwardMovement_callback() {
    socket.send(msgType_commandMoveForward);
}

function testLeftMovement_callback() {
    socket.send(msgType_commandMoveLeft);
}

function testRightMovement_callback() {
    socket.send(msgType_commandMoveRight);
}