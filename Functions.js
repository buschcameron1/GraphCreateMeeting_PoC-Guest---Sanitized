function meetNow() {
    const now = new Date();
    const endTime = new Date(now.getTime() + 30 * 60000);

    const meetingDetails = {
        subject: "Meet Now",
        start_time: now.toISOString(),
        end_time: endTime.toISOString(),
        attendees: [
            {
                emailAddress: {
                    address: "[Email Address of Attendee]",
                    name: "[Name of Attendee]"
                },
                type: "required"
            }
        ],
        organizer: "[Organizer Object ID]"
    };

    fetch('http://127.0.0.1:8080/create-event', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(meetingDetails)
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Meeting created successfully:', data);
        document.getElementById('chatBox').innerHTML += `<p>${data.body}</p>`;
    })
    .catch(error => {
        console.error('Error creating meeting:', error);
    });
}

function openScheduleModal() {
    document.getElementById('modalOverlay').style.display = 'block';
    document.getElementById('scheduleModal').style.display = 'block';
}

function closeScheduleModal() {
    document.getElementById('modalOverlay').style.display = 'none';
    document.getElementById('scheduleModal').style.display = 'none';
}

function scheduleMeeting() {
    const date = document.getElementById('meetingDate').value;
    const time = document.getElementById('meetingTime').value;
    const meetingDetails = {
            subject: "Scheduled Meeting",
            start_time: new Date(`${date}T${time}`).toISOString(),
            end_time: new Date(new Date(`${date}T${time}`).getTime() + 30 * 60000).toISOString(),
            attendees: [
                {
                    emailAddress: {
                        address: "[Email Address of Attendee]",
                        name: "[Name of Attendee]"
                    },
                    type: "required"
                }
            ],
            organizer: "[Organizer Object ID]"
        };
    if (date && time) {
        fetch('http://127.0.0.1:8080/create-event', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(meetingDetails)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            console.log('Meeting created successfully:', data);
            document.getElementById('chatBox').innerHTML += `<p>Meeting scheduled for ${date} at ${time}, please check your calendar to confirm reciept</p>`;
            closeScheduleModal();
        })
    } else {
        alert("Please select both date and time.");
    }
}