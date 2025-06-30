from odoo import models, fields, api
from datetime import datetime


class TodoTask(models.Model):
    _name = 'todo.task'
    _description = 'Todo Task'
    _inherit = ['mail.thread', 'mail.activity.mixin']

    name = fields.Char('Task Name', required=True, tracking=True)
    description = fields.Text('Description')
    start_time = fields.Datetime('Start Time')
    end_time = fields.Datetime('End Time')
    state = fields.Selection([
        ('1_draft', 'Draft'),
        ('2_in_progress', 'In Progress'),
        ('3_done', 'Done')
    ], string='Status', default='1_draft', required=True, tracking=True)
    time_spent = fields.Float('Time Spent (Hours)', compute='_compute_time_spent', store=True)
    active = fields.Boolean('Active', default=True)
    user_id = fields.Many2one('res.users', string='Assigned User', default=lambda self: self.env.user)

    @api.depends('start_time', 'end_time')
    def _compute_time_spent(self):
        for task in self:
            if task.start_time and task.end_time:
                start = fields.Datetime.from_string(task.start_time)
                end = fields.Datetime.from_string(task.end_time)
                duration = end - start
                task.time_spent = duration.total_seconds() / 3600.0
            else:
                task.time_spent = 0.0

    def action_start_task(self):
        self.write({
            'state': '2_in_progress',
            'start_time': fields.Datetime.now()
        })

    def action_complete_task(self):
        self.write({
            'state': '3_done',
            'end_time': fields.Datetime.now()
        })

    def action_reset_draft(self):
        self.write({
            'state': '1_draft',
            'start_time': False,
            'end_time': False,
            'time_spent': 0.0
        })
