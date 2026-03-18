import { useState } from 'react';
import './App.css';

interface Payment {
  amount: number;
  paid_at: string;
}

interface InvoiceDetail {
  id: number;
  name: string;
  amount: number;
  currency: string;
  status: string;
  issued_at: string;
  due_at: string;
  payments: Payment[];
}

function App() {

  // State for the Modal and Form Data Add Invocie and Payment
  const [invoiceModal, setInvoiceModal] = useState(false);
  const [paymentModal, setPaymentModal] = useState(false);


  const [invoiceFormData, setInvoiceForm] = useState({ 
    name: '',
    amount: '', 
    currency: 'USD', 
    issueDate: '',
    dueDate: '', 
    status: 'PENDING' 
  });
  
  const [paymentFormData, setPaymentForm] = useState({ 
    invoiceId: '', 
    amount: '' 
  });


  // State for Get Inovice and Payment
  const [searchId, setSearchId] = useState('');
  const [selectedInvoice, setSelectedInvoice] = useState<InvoiceDetail | null>(null);
  const [loading, setLoading] = useState(false);
  
  


  {/* Send Invoice Logic */}
  const handleInvoiceConfirm = async () => {
    try {
      const response = await fetch("http://localhost:8080/api/invoices/", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: invoiceFormData.name,
          amount: Number(invoiceFormData.amount),
          currency: invoiceFormData.currency,
          issued_at: invoiceFormData.issueDate,
          due_at: invoiceFormData.dueDate,
          status: invoiceFormData.status,
        }),
      });

      if (response.ok) {
        alert("Invoice Created!");
        setInvoiceModal(false);
        setInvoiceForm({ name: '', amount: '', currency: 'USD', issueDate: '', dueDate: '', status: 'PENDING' });
      }
    } catch (error) {
      alert("Error connecting to backend");
    }
  };

  {/* Send Payment Logic */}
  const handlePaymentConfirm = async () => {
    if (!paymentFormData.invoiceId || !paymentFormData.amount) {
      alert("Please enter both an Invoice ID and an Amount");
      return;
    }

    try {
      const response = await fetch(`http://localhost:8080/api/invoices/${paymentFormData.invoiceId}/payments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          amount: Number(paymentFormData.amount),
        }),
      });

      if (response.ok) {
        alert("Payment Recorded!");
        setPaymentModal(false);
        setPaymentForm({ invoiceId: '', amount: '' });
      } else {
        const errorData = await response.json();
        alert(`Error: ${errorData.message}`);
      }
    } catch (error) {
      alert("Network Error");
    }
  };

  const handleInvoiceSearch = async () => {
    if (!searchId) return;
    setLoading(true);
    try {
      const response = await fetch(`http://localhost:8080/api/invoices/${searchId}`);
      const result = await response.json();
      
      if (response.ok) {
        setSelectedInvoice(result.data); 
      } else {
        alert("Invoice not found");
        setSelectedInvoice(null);
      }
    } catch (error) {
      alert("Error fetching invoice details");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container">
      <h1>eCapital Portal</h1>
      
      <div style={{ display: 'flex', gap: '10px', justifyContent: 'center' }}>
        <button onClick={() => setInvoiceModal(true)}>Create Invoice</button>
        <button onClick={() => setPaymentModal(true)}>Record Payment</button>
      </div>

      {/* Create Invoice Modal */}
      {invoiceModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <h2>New Invoice</h2>
            <label>Name</label>
            <input type="string" onChange={(e) => setInvoiceForm({...invoiceFormData, name: e.target.value})} />
            
            <label>Amount</label>
            <input type="number" onChange={(e) => setInvoiceForm({...invoiceFormData, amount: e.target.value})} />
            
            <label>Currency</label>
            <select onChange={(e) => setInvoiceForm({...invoiceFormData, currency: e.target.value})}>
              <option value="USD">USD</option>
              <option value="CAD">CAD</option>
            </select>

            <label>Issue Date</label>
            <input type="date" onChange={(e) => setInvoiceForm({...invoiceFormData, issueDate: e.target.value})} />

            <label>Due Date</label>
            <input type="date" onChange={(e) => setInvoiceForm({...invoiceFormData, dueDate: e.target.value})} />

            <div style={{ marginTop: '20px' }}>
              <button onClick={handleInvoiceConfirm} style={{ background: 'var(--accent)', color: 'white' }}>Confirm</button>
              <button onClick={() => setInvoiceModal(false)} style={{ marginLeft: '10px' }}>Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Record Payment Modal */}
      {paymentModal && (
        <div className="modal-overlay">
          <div className="modal-content">
            <h2>Record Payment</h2>
            <label>Invoice ID</label>
            <input type="number" value={paymentFormData.invoiceId} onChange={(e) => setPaymentForm({...paymentFormData, invoiceId: e.target.value})} />
            
            <label>Amount</label>
            <input type="number" value={paymentFormData.amount} onChange={(e) => setPaymentForm({...paymentFormData, amount: e.target.value})} />

            <div style={{ marginTop: '20px' }}>
              <button onClick={handlePaymentConfirm} style={{ background: 'var(--accent)', color: 'white' }}>Confirm</button>
              <button onClick={() => setPaymentModal(false)} style={{ marginLeft: '10px' }}>Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Search for Invoice Modal */}
      <div className="search-section" style={{ margin: '30px 0', textAlign: 'center' }}>
        <h2 className="search-bar-title">Invoice Lookup</h2>

        <input 
          type="number" 
          placeholder="Enter Invoice ID to view details..." 
          value={searchId}
          onChange={(e) => setSearchId(e.target.value)}
          style={{ padding: '10px', width: '250px', borderRadius: '4px 0 0 4px', border: '1px solid #ccc' }}
        />
        <button 
          onClick={handleInvoiceSearch}
          style={{ padding: '10px 20px', borderRadius: '0 4px 4px 0', cursor: 'pointer' }}
        >
          {loading ? 'Searching...' : 'Search'}
        </button>
      </div>

      {/* Invoice Search Result Modal */}
      {selectedInvoice && (
      <div className="modal-overlay">
        <div className="modal-content" style={{ textAlign: 'left', minWidth: '400px' }}>
          <h2>Invoice Details</h2>
          <p><strong>ID:</strong> {selectedInvoice.id}</p>
          <p><strong>Customer:</strong> {selectedInvoice.name}</p>
          <p>
            <strong>Total: </strong> 
              {selectedInvoice.amount.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
              <strong> </strong> {selectedInvoice.currency}  
          </p>
          <p><strong>Status:</strong> {selectedInvoice.status}</p>

          <hr />
          <h3>Payments</h3>
          {selectedInvoice.payments?.length > 0 ? (
            selectedInvoice.payments.map((p, i) => (
              <div key={i} style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>{p.paid_at.split('T')[0]}</span>
                <strong>+ {p.amount}</strong>
              </div>
            ))
          ) : (
            <p>No payments recorded.</p>
          )}

          <button 
            onClick={() => setSelectedInvoice(null)} 
            style={{ marginTop: '20px', background: 'var(--accent)' }}
          >
            Close
          </button>
        </div>
      </div>
    )}
    
    </div>

    
  );
}

export default App;