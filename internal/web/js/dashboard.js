document.addEventListener('DOMContentLoaded', () => {
    const createEventForm = document.getElementById('createEventForm');
    const eventList = document.getElementById('eventList');

    // Fetch and display existing events
    const fetchEvents = async () => {
        try {
            const response = await fetch('/api/events'); // Use relative path
            if (!response.ok) {
                throw new Error(`Failed to fetch events: ${response.status} ${response.statusText}`);
            }
            const data = await response.json();

            // Clear existing list items
            eventList.innerHTML = '';

            // Add new list items for each event
            data.events.forEach(event => { // Access events array
                const listItem = document.createElement('li');
                const startTime = new Date(event.StartTime).toLocaleString();
                const endTime = new Date(event.EndTime).toLocaleString();
                listItem.textContent = `${event.Title} - ${startTime} to ${endTime} - Attendees: ${event.Attendees.join(', ')}`;
                eventList.appendChild(listItem);
            });

        } catch (error) {
            console.error('Error fetching events:', error);
            alert(`Failed to fetch events: ${error.message}`);
        }
    };

    // Create Event Form Submission
    createEventForm.addEventListener('submit', async (event) => {
        event.preventDefault();

        const formData = new FormData(createEventForm);
        const csrfToken = formData.get('csrf_token'); //Get CSRF

        const attendeesInput = formData.get('attendees');
        const attendeesArray = attendeesInput ? attendeesInput.split(',').map(email => email.trim()) : [];


        const eventData = {
            title: formData.get('title'),
            description: formData.get('description'),
            start_time: new Date(formData.get('start_time')).toISOString(), // Convert to ISO string
            end_time: new Date(formData.get('end_time')).toISOString(),     // Convert to ISO string
            attendees: attendeesArray, // Use the array of emails
        };
		console.log(eventData);

        try {
            const response = await fetch('/api/events', {  // Use relative path
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
					'X-CSRF-Token': csrfToken, // Add CSRF header
                },
                body: JSON.stringify(eventData),
            });

            if (!response.ok) {
                const errorData = await response.json(); //Try to get error.
                throw new Error(`Failed to create event: ${response.status} ${response.statusText} - ${errorData.message}`);
            }


            // Fetch and display updated event list
            await fetchEvents();
            createEventForm.reset();

        } catch (error) {
            console.error('Error creating event:', error);
            alert(`Failed to create event: ${error.message}`);
        }
    });

     fetchEvents(); // Initial fetch
});