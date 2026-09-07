
import { confirmAlert } from 'react-confirm-alert'; // Import


export default {
    confirmation: (title, cbdone, cbcancel) => confirmAlert({
        title: title,
        message: 'Are you sure to do this.',
        buttons: [
          {
            label: 'Yes',
            onClick: cbdone
          },
          {
            label: 'No',
            onClick: cbcancel
          }
        ]
      }),
    get: (endpoint: string, cbdone: (data: any) => void, cbcancel: () => void = () => {{}}) => 
        fetch(endpoint, {
            method: "GET",
            headers: {
                "Accept": "application/json"
            }
        })
        .then(response => {
            return response.json()
        })
        .then(data => cbdone(data))
        .then(_ => cbcancel)
};